namespace EmailService.Core.Exceptions;

/// <summary>
/// Represents an exception that is thrown when an email address is invalid.
/// </summary>
public class InvalidEmailException : Exception
{
    /// <summary>
    /// Initializes a new instance of the <see cref="InvalidEmailException"/> class.
    /// </summary>
    /// <param name="email">The invalid email address.</param>
    public InvalidEmailException(string email)
        : base($"Invalid email address: {email}")
    {
    }
}