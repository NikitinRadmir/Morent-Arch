namespace EmailService.Core.Exceptions;

public class ProviderDeliveryException : Exception
{
    public ProviderDeliveryException(string message, Exception inner) : base(message, inner) { }
}