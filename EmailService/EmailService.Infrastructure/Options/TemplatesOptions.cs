namespace EmailService.Infrastructure.Options;

/// <summary>
/// Configuration options for email template rendering.
/// </summary>
public class TemplatesOptions
{
    /// <summary>
    /// Base directory path where templates are stored.
    /// </summary>
    public string BasePath { get; set; } = "./Templates";

    /// <summary>
    /// Default file extension for template files.
    /// </summary>
    public string DefaultExtension { get; set; } = ".html";

    /// <summary>
    /// Template cache duration in seconds.
    /// </summary>
    public int CacheDurationSeconds { get; set; } = 3600;
}