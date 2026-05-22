namespace EmailService.Worker.Configuration;

/// <summary>
/// Configuration options for the email worker.
/// </summary>
public class WorkerOptions
{
    /// <summary>
    /// Maximum number of concurrent email processing tasks.
    /// </summary>
    public int MaxConcurrency { get; set; } = 4;

    /// <summary>
    /// Maximum number of retry attempts for failed operations.
    /// </summary>
    public int MaxRetries { get; set; } = 3;

    /// <summary>
    /// Timeout for graceful shutdown of the worker.
    /// </summary>
    public TimeSpan ShutdownTimeout { get; set; } = TimeSpan.FromSeconds(30);

    /// <summary>
    /// Path to the email templates directory.
    /// </summary>
    public string TemplatesPath { get; set; } = "./Templates";
}