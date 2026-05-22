namespace EmailService.Core.Exceptions;

/// <summary>
/// Represents an exception that occurs when email delivery through a provider fails.
/// </summary>
public class ProviderDeliveryException : Exception
{
    /// <summary>
    /// Initializes a new instance of the <see cref="ProviderDeliveryException"/> class.
    /// </summary>
    /// <param name="message">The exception message.</param>
    /// <param name="inner">The inner exception.</param>
    public ProviderDeliveryException(string message, Exception inner)
        : base(message, inner)
    {
    }
}