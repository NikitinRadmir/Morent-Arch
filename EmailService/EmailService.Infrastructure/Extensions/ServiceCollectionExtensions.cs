using EmailService.Core.Contracts;
using EmailService.Core.Extensions;
using EmailService.Infrastructure.Consumers;
using EmailService.Infrastructure.Options;
using EmailService.Infrastructure.Persistence;
using EmailService.Infrastructure.Providers;
using EmailService.Infrastructure.Rendering;
using MassTransit;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace EmailService.Infrastructure.Extensions;

/// <summary>
/// Provides extension methods for registering infrastructure services.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Registers infrastructure services for the email system.
    /// </summary>
    /// <param name="services">The service collection.</param>
    /// <param name="config">Application configuration.</param>
    /// <returns>The updated service collection.</returns>
    public static IServiceCollection AddEmailInfrastructure(
        this IServiceCollection services,
        IConfiguration config)
    {
        services.AddEmailCore();

        services.Configure<TemplatesOptions>(config.GetSection("Templates"));
        services.Configure<ResendOptions>(config.GetSection("Resend"));
        services.Configure<RabbitMqOptions>(config.GetSection("RabbitMq"));

        services.AddSingleton<IMemoryCache, MemoryCache>();
        services.AddSingleton<ITemplateRenderer, ScribanTemplateRenderer>();

        // 🔹 SMTP Provider
        services.Configure<SmtpOptions>(config.GetSection("Smtp"));
        services.AddSingleton<IEmailProvider, SmtpEmailProvider>();

        services.AddDbContext<EmailDbContext>(opt =>
            opt.UseNpgsql(config.GetConnectionString("EmailDb")));

        services.AddScoped<IEmailStatusStore, EfEmailStatusStore>();

        services.AddMassTransit(x =>
        {
            x.AddConsumer<EmailConsumer>();

            x.UsingRabbitMq((context, cfg) =>
            {
                cfg.Host(config["RabbitMq:Host"], h =>
                {
                    h.Username(config["RabbitMq:User"]);
                    h.Password(config["RabbitMq:Pass"]);
                });

                cfg.ConfigureEndpoints(context);
            });
        });

        return services;
    }
}

/// <summary>
/// Represents RabbitMQ connection settings.
/// </summary>
public class RabbitMqOptions
{
    /// <summary>
    /// RabbitMQ host address.
    /// </summary>
    public string Host { get; set; } = "localhost";

    /// <summary>
    /// RabbitMQ username.
    /// </summary>
    public string User { get; set; } = "guest";

    /// <summary>
    /// RabbitMQ password.
    /// </summary>
    public string Pass { get; set; } = "guest";
}