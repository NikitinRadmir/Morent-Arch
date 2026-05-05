namespace EmailService.Infrastructure.Options;

public class TemplatesOptions
{
    public string BasePath { get; set; } = "./Templates";
    public string DefaultExtension { get; set; } = ".html";
    public int CacheDurationSeconds { get; set; } = 3600;
}