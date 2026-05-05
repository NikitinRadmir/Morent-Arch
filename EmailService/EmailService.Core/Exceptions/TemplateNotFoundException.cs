namespace EmailService.Core.Exceptions;

public class TemplateNotFoundException : Exception
{
    public TemplateNotFoundException(string key) : base($"Template '{key}' not found.") { }
}