namespace EmailService.Core.Exceptions;

/// <summary>
/// Represents an exception that is thrown when an email template cannot be found.
/// </summary>
public class TemplateNotFoundException : Exception
{
    /// <summary>
    /// Initializes a new instance of the <see cref="TemplateNotFoundException"/> class.
    /// </summary>
    /// <param name="key">The template key.</param>
    public TemplateNotFoundException(string key)
        : base($"Template '{key}' not found.")
    {
    }
}