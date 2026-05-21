using EmailService.Core.Contracts;
using EmailService.Core.Models;
using MailKit.Net.Smtp;
using MailKit.Security;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using MimeKit;

namespace EmailService.Infrastructure.Providers;

public class SmtpEmailProvider : IEmailProvider
{
    private readonly IConfiguration _config;
    private readonly ILogger<SmtpEmailProvider> _logger;

    public SmtpEmailProvider(IConfiguration config, ILogger<SmtpEmailProvider> logger)
    {
        _config = config;
        _logger = logger;
    }

    public async Task SendAsync(RenderedEmail email, CancellationToken ct)
    {
        var message = new MimeMessage();

        message.From.Add(new MailboxAddress(
            _config["Smtp:FromName"] ?? "AutoRental",
            _config["Smtp:FromEmail"]));

        message.To.Add(new MailboxAddress("", email.To));
        message.Subject = email.Subject;

        var bodyBuilder = new BodyBuilder { HtmlBody = email.HtmlContent };
        message.Body = bodyBuilder.ToMessageBody();

        using var client = new SmtpClient();

        var host = _config["Smtp:Host"] ?? throw new InvalidOperationException("SMTP Host not configured");
        var port = int.TryParse(_config["Smtp:Port"], out var p) ? p : 587;
        var useSsl = bool.TryParse(_config["Smtp:UseSsl"], out var ssl) && ssl;
        var secureOption = useSsl ? SecureSocketOptions.StartTls : SecureSocketOptions.Auto;

        await client.ConnectAsync(host, port, secureOption, ct);

        var username = _config["Smtp:Username"];
        var password = _config["Smtp:Password"];

        if (!string.IsNullOrEmpty(username))
        {
            await client.AuthenticateAsync(username, password, ct);
        }

        await client.SendAsync(message, ct);
        await client.DisconnectAsync(true, ct);

        _logger.LogInformation("Email sent via SMTP to {To}", email.To);
    }
}