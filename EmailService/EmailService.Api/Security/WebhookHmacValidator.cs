using System.Security.Cryptography;
using System.Text;
using EmailService.Infrastructure.Options;
using Microsoft.Extensions.Options;

namespace EmailService.Infrastructure.Security;

/// <summary>
/// Validates webhook requests using HMAC signature verification.
/// </summary>
public class WebhookHmacValidator
{
    private readonly string _secret;

    /// <summary>
    /// Initializes a new instance of the <see cref="WebhookHmacValidator"/> class.
    /// </summary>
    /// <param name="options">Configuration options for Resend integration.</param>
    /// <exception cref="InvalidOperationException">
    /// Thrown when neither WebhookSecret nor ApiKey is configured.
    /// </exception>
    public WebhookHmacValidator(IOptions<ResendOptions> options)
    {
        _secret = options.Value.WebhookSecret
                  ?? options.Value.ApiKey
                  ?? throw new InvalidOperationException(
                      "Resend:WebhookSecret or ApiKey must be configured");
    }

    /// <summary>
    /// Verifies the HMAC signature of an incoming webhook payload.
    /// </summary>
    /// <param name="signature">The signature provided in the webhook request headers.</param>
    /// <param name="payload">The webhook payload object.</param>
    /// <returns><c>true</c> if the signature is valid or validation is skipped; otherwise <c>false</c>.</returns>
    /// <remarks>
    /// If no secret or signature is configured, verification is bypassed (development mode).
    /// </remarks>
    public bool Verify(string? signature, object? payload)
    {
        if (string.IsNullOrEmpty(_secret) || string.IsNullOrEmpty(signature))
            return true;

        if (payload == null)
            return false;

        var json = System.Text.Json.JsonSerializer.Serialize(payload);
        var expected = ComputeHmacSha256(json, _secret);

        return CryptographicOperations.FixedTimeEquals(
            Encoding.UTF8.GetBytes(signature),
            Encoding.UTF8.GetBytes(expected));
    }

    /// <summary>
    /// Computes a SHA256 HMAC hash for the given message using the specified secret.
    /// </summary>
    /// <param name="message">The message to hash.</param>
    /// <param name="secret">The secret key used for hashing.</param>
    /// <returns>Hexadecimal string representation of the HMAC hash.</returns>
    private static string ComputeHmacSha256(string message, string secret)
    {
        using var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(secret));
        var hash = hmac.ComputeHash(Encoding.UTF8.GetBytes(message));
        return Convert.ToHexString(hash).ToLowerInvariant();
    }
}