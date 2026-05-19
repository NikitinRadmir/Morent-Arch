using System.Net.Http.Json;
using EmailService.Core.Contracts;
using EmailService.Core.Exceptions;
using EmailService.Core.Models;
using EmailService.Infrastructure.Options;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace EmailService.Infrastructure.Providers;

/// <summary>
/// Email provider implementation for the Resend API.
/// </summary>
public class ResendEmailProvider : IEmailProvider
{
    private readonly HttpClient _httpClient;
    private readonly ResendOptions _options;
    private readonly IEmailStatusStore _store;
    private readonly ILogger<ResendEmailProvider> _logger;

    /// <summary>
    /// Initializes a new instance of the <see cref="ResendEmailProvider"/> class.
    /// </summary>
    /// <param name="httpClient">HTTP client for API requests.</param>
    /// <param name="options">Resend configuration options.</param>
    /// <param name="store">Email status store.</param>
    /// <param name="logger">Logger instance.</param>
    public ResendEmailProvider(
        HttpClient httpClient,
        IOptions<ResendOptions> options,
        IEmailStatusStore store,
        ILogger<ResendEmailProvider> logger)
    {
        _httpClient = httpClient;
        _options = options.Value;
        _store = store;
        _logger = logger;
    }

    /// <summary>
    /// Sends a rendered email through the Resend API.
    /// </summary>
    /// <param name="email">Rendered email message.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <exception cref="ProviderDeliveryException">
    /// Thrown when the email delivery fails.
    /// </exception>
    public async Task SendAsync(RenderedEmail email, CancellationToken ct = default)
    {
        var payload = new
        {
            from = $"{_options.FromName} <{_options.FromEmail}>",
            to = new[] { email.To },
            subject = email.Subject,
            html = email.HtmlContent,
            headers = new Dictionary<string, string>
            {
                ["X-Correlation-Id"] = email.CorrelationId
            }
        };

        try
        {
            var response = await _httpClient.PostAsJsonAsync(
                "emails",
                payload,
                ct);

            var content = await response.Content.ReadAsStringAsync(ct);

            if (!response.IsSuccessStatusCode)
            {
                _logger.LogError(
                    "Resend API error {StatusCode}: {Content}",
                    response.StatusCode,
                    content);

                throw new ProviderDeliveryException(
                    $"Resend returned {(int)response.StatusCode}",
                    null!);
            }

            _logger.LogInformation(
                "Resend accepted {CorrelationId}. Provider ID: {ResendId}",
                email.CorrelationId,
                content);

            await _store.MarkStatusAsync(
                email.CorrelationId,
                EmailStatus.Sent,
                ct: ct);
        }
        catch (HttpRequestException ex)
        {
            _logger.LogError(
                ex,
                "Network error sending to Resend {CorrelationId}",
                email.CorrelationId);

            throw new ProviderDeliveryException(
                "Resend network error",
                ex);
        }
    }
}