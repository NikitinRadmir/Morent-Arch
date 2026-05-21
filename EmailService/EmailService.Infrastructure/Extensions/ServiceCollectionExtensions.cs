using EmailService.Core.Contracts;
using EmailService.Core.Extensions;
using EmailService.Infrastructure.Messaging;
using EmailService.Infrastructure.Options;
using EmailService.Infrastructure.Persistence;
using EmailService.Infrastructure.Providers;
using EmailService.Infrastructure.Queue;
using EmailService.Infrastructure.Rendering;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Caching.Memory;        // ← ДОБАВИТЬ (для IMemoryCache, MemoryCache)
using Microsoft.Extensions.Configuration;          // ← ДОБАВИТЬ (для IConfiguration)
using Microsoft.Extensions.DependencyInjection;

namespace EmailService.Infrastructure.Extensions;

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddEmailInfrastructure(this IServiceCollection services, IConfiguration config)
    {
        services.AddEmailCore();

        // Options
        services.Configure<TemplatesOptions>(config.GetSection("Templates"));
        services.Configure<SendGridOptions>(config.GetSection("SendGrid"));
        services.Configure<KafkaOptions>(config.GetSection("Kafka"));

        // Queue
        services.AddSingleton<IEmailQueue, ChannelEmailQueue>();

        // Rendering
        services.AddSingleton<IMemoryCache, MemoryCache>();
        services.AddSingleton<ITemplateRenderer, ScribanTemplateRenderer>();

        // Persistence
        services.AddDbContext<EmailDbContext>(opt =>
            opt.UseSqlite(config.GetConnectionString("EmailDb") ?? "Data Source=email.db"));
        services.AddScoped<IEmailStatusStore, EfEmailStatusStore>();

        // Provider: SendGrid при наличии ключа, иначе dev-лог (без реальной доставки)
        var sendGridKey = config.GetSection("SendGrid")["ApiKey"];
        if (string.IsNullOrWhiteSpace(sendGridKey))
        {
            services.AddScoped<IEmailProvider, DevLoggingEmailProvider>();
        }
        else
        {
            services.AddScoped<IEmailProvider, SendGridEmailProvider>();
        }

        return services;
    }

    public static IServiceCollection AddEmailKafkaConsumer(this IServiceCollection services)
    {
        services.AddScoped<MorentEmailIngestion>();
        services.AddHostedService<KafkaEmailConsumer>();
        return services;
    }
}