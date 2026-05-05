namespace EmailService.Worker.Configuration;

public class WorkerOptions
{
    public int MaxConcurrency { get; set; } = 4;
    public int MaxRetries { get; set; } = 3; 
    public TimeSpan ShutdownTimeout { get; set; } = TimeSpan.FromSeconds(30);
    public string TemplatesPath { get; set; } = "./Templates";
}