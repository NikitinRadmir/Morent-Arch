using EmailService.Core.Contracts;
using EmailService.Core.Extensions;
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

        // Queue
        services.AddSingleton<IEmailQueue, ChannelEmailQueue>();

        // Rendering
        services.AddSingleton<IMemoryCache, MemoryCache>();
        services.AddSingleton<ITemplateRenderer, ScribanTemplateRenderer>();

        // Persistence
        services.AddDbContext<EmailDbContext>(opt =>
            opt.UseSqlite(config.GetConnectionString("EmailDb") ?? "Data Source=email.db"));
        services.AddScoped<IEmailStatusStore, EfEmailStatusStore>();

        // Provider
        services.AddScoped<IEmailProvider, SendGridEmailProvider>();

        return services;
    }
}