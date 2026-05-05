using EmailService.Api.Endpoints;
using EmailService.Api.Validators;
using EmailService.Infrastructure.Extensions;
using EmailService.Infrastructure.Options;
using EmailService.Infrastructure.Persistence;
using EmailService.Infrastructure.Security;
using FluentValidation;
using Microsoft.AspNetCore.Mvc;
using Serilog;
using Serilog.Events;
using FluentValidation.AspNetCore;                   

Log.Logger = new LoggerConfiguration()
    .MinimumLevel.Information()
    .MinimumLevel.Override("Microsoft", LogEventLevel.Warning)
    .MinimumLevel.Override("System", LogEventLevel.Warning)
    .WriteTo.Console(outputTemplate: "[{Timestamp:HH:mm:ss} {Level:u3}] {Message:lj}{NewLine}{Exception}")
    .Enrich.FromLogContext()
    .CreateLogger();

try
{
    Log.Information("Starting EmailService.Api");

    var builder = WebApplication.CreateBuilder(args);

    builder.Configuration.AddEnvironmentVariables();
    builder.Services.Configure<TemplatesOptions>(builder.Configuration.GetSection("Templates"));
    builder.Services.Configure<SendGridOptions>(builder.Configuration.GetSection("SendGrid"));

    builder.Host.UseSerilog();

    builder.Services.AddControllers();
    builder.Services.AddFluentValidationAutoValidation();
    builder.Services.AddValidatorsFromAssemblyContaining<EmailRequestValidator>();

    builder.Services.Configure<ApiBehaviorOptions>(opt =>
    {
        opt.InvalidModelStateResponseFactory = ctx =>
        {
            var errors = ctx.ModelState
                .Where(e => e.Value?.Errors.Any() == true)
                .Select(e => new { field = e.Key, errors = e.Value!.Errors.Select(err => err.ErrorMessage) })
                .ToList();

            return new BadRequestObjectResult(new
            {
                type = "https://tools.ietf.org/html/rfc7231#section-6.5.1",
                title = "Validation Failed",
                status = 400,
                errors
            });
        };
    });

    builder.Services.AddEmailInfrastructure(builder.Configuration);
    builder.Services.AddSingleton<WebhookHmacValidator>();

    builder.Services.AddEndpointsApiExplorer();
    builder.Services.AddSwaggerGen(c =>
    {
        c.SwaggerDoc("v1", new() { Title = "Email Service API", Version = "v1" });
    });

    var app = builder.Build();

    app.UseSerilogRequestLogging();
    app.UseHttpsRedirection();
    app.UseAuthorization();

    if (app.Environment.IsDevelopment())
    {
        app.UseSwagger();
        app.UseSwaggerUI();
    }

    app.MapIngestionEndpoints();
    app.MapStatusEndpoints();
    app.MapWebhookEndpoints();

    app.MapGet("/health", () => Results.Ok(new { status = "healthy" }));

    using (var scope = app.Services.CreateScope())
    {
        var db = scope.ServiceProvider.GetRequiredService<EmailDbContext>();
        await db.Database.EnsureCreatedAsync();
    }

    app.Run();
}
catch (Exception ex)
{
    Log.Fatal(ex, "Api terminated unexpectedly");
    throw;
}
finally
{
    Log.CloseAndFlush();
}