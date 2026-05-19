using EmailService.Core.Contracts;
using EmailService.Core.Models;
using EmailService.Worker.Configuration;
using EmailService.Worker.Consumers;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace EmailService.Worker;

/// <summary>
/// Background worker that processes email requests from the queue with controlled concurrency.
/// </summary>
public class Worker : BackgroundService
{
    private readonly IEmailQueue _queue;
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly WorkerOptions _options;
    private readonly ILogger<Worker> _logger;
    private readonly SemaphoreSlim _semaphore;

    /// <summary>
    /// Initializes a new instance of the <see cref="Worker"/> class.
    /// </summary>
    /// <param name="queue">Email request queue.</param>
    /// <param name="scopeFactory">Service scope factory for resolving scoped services.</param>
    /// <param name="options">Worker configuration options.</param>
    /// <param name="logger">Logger instance.</param>
    public Worker(
        IEmailQueue queue,
        IServiceScopeFactory scopeFactory,
        IOptions<WorkerOptions> options,
        ILogger<Worker> logger)
    {
        _queue = queue;
        _scopeFactory = scopeFactory;
        _options = options.Value;
        _logger = logger;
        _semaphore = new SemaphoreSlim(_options.MaxConcurrency);
    }

    /// <summary>
    /// Executes the background processing loop for email requests.
    /// </summary>
    /// <param name="stoppingToken">Cancellation token.</param>
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation(
            "Worker started with concurrency {Concurrency}",
            _options.MaxConcurrency);

        await foreach (var request in _queue.ReadAllAsync(stoppingToken))
        {
            await _semaphore.WaitAsync(stoppingToken);

            _ = Task.Run(async () =>
            {
                using var scope = _scopeFactory.CreateScope();

                try
                {
                    var dispatcher =
                        scope.ServiceProvider.GetRequiredService<EmailDispatcher>();

                    await dispatcher.ProcessAsync(request, stoppingToken);
                }
                catch (Exception ex)
                {
                    _logger.LogError(
                        ex,
                        "Unhandled error in worker task for {CorrelationId}",
                        request.CorrelationId);
                }
                finally
                {
                    _semaphore.Release();
                }
            }, stoppingToken);
        }
    }

    /// <summary>
    /// Performs graceful shutdown and waits for running tasks to complete.
    /// </summary>
    /// <param name="cancellationToken">Cancellation token.</param>
    public override async Task StopAsync(CancellationToken cancellationToken)
    {
        _logger.LogInformation("Worker stopping, waiting for tasks to complete...");

        using var cts =
            CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);

        cts.CancelAfter(_options.ShutdownTimeout);

        while (_semaphore.CurrentCount < _options.MaxConcurrency)
        {
            try
            {
                await _semaphore.WaitAsync(
                    TimeSpan.FromMilliseconds(100),
                    cts.Token);

                _semaphore.Release();
            }
            catch (OperationCanceledException)
            {
                break;
            }
        }

        _logger.LogInformation("Worker stopped");

        await base.StopAsync(cancellationToken);
    }
}