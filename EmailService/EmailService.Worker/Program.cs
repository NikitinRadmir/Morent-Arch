using EmailService.Infrastructure.Extensions;
using EmailService.Infrastructure.Options;
using EmailService.Infrastructure.Persistence;
using EmailService.Worker;
using EmailService.Worker.Configuration;
using EmailService.Worker.Consumers;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Serilog;
using Serilog.Events;

var logConfig = new LoggerConfiguration()
    .MinimumLevel.Information()
    .MinimumLevel.Override("Microsoft", LogEventLevel.Warning)
    .MinimumLevel.Override("System", LogEventLevel.Warning)
    .WriteTo.Console(outputTemplate: "[{Timestamp:HH:mm:ss} {Level:u3}] {Message:lj}{NewLine}{Exception}")
    .Enrich.FromLogContext();

Log.Logger = logConfig.CreateLogger();

try
{
    Log.Information("Starting EmailService.Worker");

    var builder = Host.CreateApplicationBuilder(args);

    builder.Configuration.AddEnvironmentVariables();
    builder.Services.Configure<WorkerOptions>(builder.Configuration.GetSection("Worker"));
    builder.Services.Configure<TemplatesOptions>(builder.Configuration.GetSection("Templates"));
    builder.Services.Configure<SendGridOptions>(builder.Configuration.GetSection("SendGrid"));

    builder.Logging.AddSerilog(dispose: true);

    builder.Services.Configure<HostOptions>(o =>
        o.BackgroundServiceExceptionBehavior = BackgroundServiceExceptionBehavior.Ignore);

    builder.Services.AddEmailInfrastructure(builder.Configuration);
    builder.Services.AddEmailKafkaConsumer();
    builder.Services.AddScoped<EmailDispatcher>();
    builder.Services.AddHostedService<Worker>();

    var host = builder.Build();

    using (var scope = host.Services.CreateScope())
    {
        var db = scope.ServiceProvider.GetRequiredService<EmailDbContext>();
        await db.Database.EnsureCreatedAsync();
    }

    await host.RunAsync();
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