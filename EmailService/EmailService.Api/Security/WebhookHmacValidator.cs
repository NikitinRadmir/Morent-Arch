using System.Security.Cryptography;
using System.Text;
using Microsoft.Extensions.Options;
using EmailService.Infrastructure.Options;

namespace EmailService.Infrastructure.Security;

public class WebhookHmacValidator
{
    private readonly string _secret;

    public WebhookHmacValidator(IOptions<SendGridOptions> options)
    {
        _secret = options.Value.ApiKey;
    }

    public bool Verify(string signature, object? payload)
    {
        if (string.IsNullOrEmpty(signature) || payload == null) return false;

        var json = System.Text.Json.JsonSerializer.Serialize(payload);
        var expected = ComputeHmacSha256(json, _secret);

        return CryptographicOperations.FixedTimeEquals(
            Encoding.UTF8.GetBytes(signature),
            Encoding.UTF8.GetBytes(expected));
    }

    private static string ComputeHmacSha256(string message, string secret)
    {
        using var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(secret));
        var hash = hmac.ComputeHash(Encoding.UTF8.GetBytes(message));
        return Convert.ToHexString(hash).ToLowerInvariant();
    }
}