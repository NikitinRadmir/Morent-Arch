namespace EmailService.Core.Exceptions;

public class InvalidEmailException : Exception
{
    public InvalidEmailException(string email) : base($"Invalid email address: {email}") { }
}
