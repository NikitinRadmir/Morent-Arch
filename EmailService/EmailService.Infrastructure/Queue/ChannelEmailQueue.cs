using System.Threading.Channels;
using EmailService.Core.Contracts;
using EmailService.Core.Models;

namespace EmailService.Infrastructure.Queue;

/// <summary>
/// Channel-based implementation of <see cref="IEmailQueue"/>.
/// </summary>
public class ChannelEmailQueue : IEmailQueue
{
    private readonly Channel<EmailRequest> _channel;

    /// <summary>
    /// Initializes a new instance of the <see cref="ChannelEmailQueue"/> class.
    /// </summary>
    /// <param name="capacity">Maximum queue capacity.</param>
    public ChannelEmailQueue(int capacity = 1000)
    {
        _channel = Channel.CreateBounded<EmailRequest>(
            new BoundedChannelOptions(capacity)
            {
                FullMode = BoundedChannelFullMode.Wait,
                SingleWriter = false,
                SingleReader = false
            });
    }

    /// <summary>
    /// Adds an email request to the queue asynchronously.
    /// </summary>
    /// <param name="request">Email request to enqueue.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>A task representing the asynchronous enqueue operation.</returns>
    public ValueTask EnqueueAsync(
        EmailRequest request,
        CancellationToken ct = default) =>
        _channel.Writer.WriteAsync(request, ct);

    /// <summary>
    /// Reads all email requests from the queue asynchronously.
    /// </summary>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>An asynchronous stream of email requests.</returns>
    public IAsyncEnumerable<EmailRequest> ReadAllAsync(
        CancellationToken ct = default) =>
        _channel.Reader.ReadAllAsync(ct);
}