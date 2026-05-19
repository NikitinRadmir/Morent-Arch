using DotNetEnv;
using EmailService.Api.Endpoints;
using EmailService.Infrastructure.Extensions;
using EmailService.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;
using Serilog;
using Serilog.Events;

if (Environment.GetEnvironmentVariable("ASPNETCORE_ENVIRONMENT") == "Development")
{
    Env.Load();
}

Log.Logger = new LoggerConfiguration()
    .MinimumLevel.Information()
    .MinimumLevel.Override("Microsoft", LogEventLevel.Warning)
    .MinimumLevel.Override("Microsoft.EntityFrameworkCore", LogEventLevel.Warning)
    .MinimumLevel.Override("System", LogEventLevel.Warning)
    .MinimumLevel.Override("MassTransit", LogEventLevel.Information)

    .Enrich.FromLogContext()
    .Enrich.WithMachineName()
    .Enrich.WithProcessId()
    .Enrich.WithProcessName()
    .Enrich.WithEnvironmentName()
    .Enrich.WithProperty("Application", "EmailService.Api")

    .WriteTo.Console(
        outputTemplate:
        "[{Timestamp:HH:mm:ss} {Level:u3}] {Message:lj}{NewLine}{Exception}")

    .WriteTo.File(
        path: "logs/api-.log",
        rollingInterval: RollingInterval.Day,
        retainedFileCountLimit: 7,
        outputTemplate:
        "{Timestamp:yyyy-MM-dd HH:mm:ss.fff zzz} [{Level:u3}] {Message:lj}{NewLine}{Exception}")

    .CreateLogger();

try
{
    Log.Information(
        "Starting EmailService.Api on {Environment} at {MachineName}",
        Environment.GetEnvironmentVariable("ASPNETCORE_ENVIRONMENT") ?? "Unknown",
        Environment.MachineName);

    var builder = WebApplication.CreateBuilder(args);

    builder.Host.UseSerilog();

    builder.Configuration.AddEnvironmentVariables();

    builder.Services.AddEmailInfrastructure(builder.Configuration);

    builder.Services.AddEndpointsApiExplorer();
    builder.Services.AddSwaggerGen();

    var app = builder.Build();

    app.UseSerilogRequestLogging(opts =>
    {
        opts.IncludeQueryInRequestPath = true;

        opts.MessageTemplate =
            "HTTP {RequestMethod} {RequestPath} responded {StatusCode} in {Elapsed:0.0000} ms";
    });

    app.UseSwagger();

    app.UseSwaggerUI(c =>
    {
        c.SwaggerEndpoint("/swagger/v1/swagger.json", "EmailService API V1");
        c.RoutePrefix = "swagger";
    });

    app.MapIngestionEndpoints();
    app.MapStatusEndpoints();
    app.MapWebhookEndpoints();

    app.MapGet("/health", () =>
        Results.Ok(new
        {
            status = "healthy",
            service = "EmailService.Api",
            time = DateTime.UtcNow
        }));

    using (var scope = app.Services.CreateScope())
    {
        var db = scope.ServiceProvider.GetRequiredService<EmailDbContext>();

        Log.Information("Ensuring database is created...");

        await db.Database.EnsureCreatedAsync();

        Log.Information("Database ensured created");
    }

    Log.Information("EmailService.Api started successfully");

    await app.RunAsync();
}
catch (Exception ex)
{
    Log.Fatal(ex, "EmailService.Api terminated unexpectedly");
}
finally
{
    Log.CloseAndFlush();
}