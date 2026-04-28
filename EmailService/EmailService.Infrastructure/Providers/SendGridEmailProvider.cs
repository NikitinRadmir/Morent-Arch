using EmailService.Core.Contracts;
using EmailService.Core.Exceptions;
using EmailService.Core.Models;
using EmailService.Infrastructure.Options; 
using Microsoft.Extensions.Logging;     
using Microsoft.Extensions.Options;
using Polly;
using SendGrid;
using SendGrid.Helpers.Mail;

namespace EmailService.Infrastructure.Providers;

public class SendGridEmailProvider : IEmailProvider
{
    private readonly SendGridClient _client;
    private readonly SendGridOptions _options;
    private readonly IEmailStatusStore _store;
    private readonly ILogger<SendGridEmailProvider> _logger;

    public SendGridEmailProvider(
        IOptions<SendGridOptions> options,
        IEmailStatusStore store,
        ILogger<SendGridEmailProvider> logger)
    {
        _options = options.Value;
        _store = store;
        _logger = logger;
        _client = new SendGridClient(_options.ApiKey);
    }

    public async Task SendAsync(RenderedEmail email, CancellationToken ct = default)
    {
        // Политика: повтор при сетевых ошибках или 5xx/429 от SendGrid
        var retryPolicy = Policy
            .Handle<HttpRequestException>()
            .Or<InvalidOperationException>()
            .WaitAndRetryAsync(
                retryCount: 3,
                sleepDurationProvider: attempt => TimeSpan.FromSeconds(Math.Pow(2, attempt)),
                onRetry: (exception, timespan, attempt, context) =>
                    _logger.LogWarning(exception, "SendGrid retry {Attempt} after {Delay}s", attempt, timespan.TotalSeconds));

        try
        {
            var msg = new SendGridMessage
            {
                From = new EmailAddress(_options.FromEmail, _options.FromName),
                Subject = email.Subject,
                HtmlContent = email.HtmlContent,
                CustomArgs = new Dictionary<string, string> { ["correlation_id"] = email.CorrelationId }
            };
            msg.AddTo(new EmailAddress(email.To));

            // Выполняем отправку с политикой повторных попыток
            await retryPolicy.ExecuteAsync(async (token) =>
            {
                var response = await _client.SendEmailAsync(msg, token);

                if (!response.IsSuccessStatusCode)
                {
                    var body = await response.Body.ReadAsStringAsync(token);
                    // Если статус 429 или 5xx — бросаем исключение для retry
                    if ((int)response.StatusCode >= 500 || response.StatusCode == System.Net.HttpStatusCode.TooManyRequests)
                        throw new ProviderDeliveryException($"SendGrid error {(int)response.StatusCode}", null);

                    // Если 4xx (кроме 429) — это клиентская ошибка, не ретраим, но логируем
                    _logger.LogError("SendGrid client error {StatusCode}: {Body} for {CorrelationId}",
                        response.StatusCode, body, email.CorrelationId);
                }

                return response;
            }, ct);

            await _store.MarkStatusAsync(email.CorrelationId, EmailStatus.Sent, ct: ct);
        }
        catch (ProviderDeliveryException)
        {
            // Уже залогировано, пробрасываем дальше для обработки в воркере
            throw;
        }
        catch (OperationCanceledException)
        {
            // Graceful shutdown, не логируем как ошибку
            throw;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send {CorrelationId}", email.CorrelationId);
            await _store.MarkStatusAsync(email.CorrelationId, EmailStatus.Failed, ex.Message, ct);
            throw;
        }
    }
}