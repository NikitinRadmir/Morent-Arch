using DotNetEnv;
using EmailService.Infrastructure.Extensions;
using EmailService.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Serilog;
using Serilog.Events;

using AppHost = Microsoft.Extensions.Hosting.Host;

if (Environment.GetEnvironmentVariable("DOTNET_ENVIRONMENT") == "Development")
{
    Env.Load();
}

Log.Logger = new LoggerConfiguration()
    .MinimumLevel.Information()
    .MinimumLevel.Override("Microsoft", LogEventLevel.Warning)
    .MinimumLevel.Override("Microsoft.EntityFrameworkCore", LogEventLevel.Warning)
    .MinimumLevel.Override("MassTransit", LogEventLevel.Debug)

    .Enrich.FromLogContext()
    .Enrich.WithMachineName()
    .Enrich.WithProcessId()
    .Enrich.WithProcessName()
    .Enrich.WithEnvironmentName()
    .Enrich.WithProperty("Application", "EmailService.Worker")

    .WriteTo.Console(
        outputTemplate:
        "[{Timestamp:HH:mm:ss} {Level:u3}] {SourceContext} {Message:lj}{NewLine}{Exception}",
        restrictedToMinimumLevel: LogEventLevel.Information)

    .WriteTo.File(
        path: "logs/worker-.log",
        rollingInterval: RollingInterval.Day,
        retainedFileCountLimit: 7,
        outputTemplate:
        "{Timestamp:yyyy-MM-dd HH:mm:ss.fff zzz} [{Level:u3}] {SourceContext} {Message:lj}{NewLine}{Exception}",
        restrictedToMinimumLevel: LogEventLevel.Information)

    .CreateLogger();

try
{
    Log.Information(
        "Starting EmailService.Worker on {Environment} at {MachineName}",
        Environment.GetEnvironmentVariable("DOTNET_ENVIRONMENT") ?? "Unknown",
        Environment.MachineName);

    var builder = AppHost.CreateApplicationBuilder(args);

    builder.Services.AddSerilog();
    builder.Configuration.AddEnvironmentVariables();

    builder.Services.AddEmailInfrastructure(builder.Configuration);

    if (builder.Environment.IsDevelopment())
    {
        var resendKey = builder.Configuration["Resend:ApiKey"];
        var connStr = builder.Configuration.GetConnectionString("EmailDb");

        Log.Debug(
            "Config loaded: Resend:ApiKey={HasKey}, EmailDb={HasConnStr}",
            !string.IsNullOrWhiteSpace(resendKey),
            !string.IsNullOrWhiteSpace(connStr));

        var maskedKey = !string.IsNullOrWhiteSpace(resendKey)
            ? resendKey[..Math.Min(10, resendKey.Length)] + "***"
            : "(empty)";

        Log.Debug("Resend:ApiKey (masked): {Key}", maskedKey);

        if (!string.IsNullOrWhiteSpace(connStr))
        {
            var maskedConn = connStr.Replace("password123", "***");
            Log.Debug("EmailDb ConnectionString: {ConnStr}", maskedConn);
        }
    }

    builder.Services.AddHostedService<DatabaseInitializerHostedService>();

    var app = builder.Build();

    Log.Information("EmailService.Worker started successfully");

    await app.RunAsync();
}
catch (Exception ex)
{
    Log.Fatal(ex, "Worker terminated unexpectedly");
    throw;
}
finally
{
    Log.CloseAndFlush();
}

public sealed class DatabaseInitializerHostedService : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;

    public DatabaseInitializerHostedService(IServiceProvider serviceProvider)
    {
        _serviceProvider = serviceProvider;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        try
        {
            using var scope = _serviceProvider.CreateScope();

            var db = scope.ServiceProvider.GetRequiredService<EmailDbContext>();

            Log.Information("Ensuring database is created...");

            await db.Database.EnsureCreatedAsync(stoppingToken);

            Log.Information("Database ensured created");
        }
        catch (Exception ex)
        {
            Log.Fatal(ex, "Database initialization failed");
            throw;
        }
    }
}