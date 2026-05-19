using EmailService.Infrastructure.Persistence.Entities;
using Microsoft.EntityFrameworkCore;

namespace EmailService.Infrastructure.Persistence;

/// <summary>
/// Database context for email service persistence.
/// </summary>
public class EmailDbContext : DbContext
{
    /// <summary>
    /// Initializes a new instance of the <see cref="EmailDbContext"/> class.
    /// </summary>
    /// <param name="options">Database context options.</param>
    public EmailDbContext(DbContextOptions<EmailDbContext> options)
        : base(options)
    {
    }

    /// <summary>
    /// Gets the email logs table.
    /// </summary>
    public DbSet<EmailLog> EmailLogs => Set<EmailLog>();

    /// <summary>
    /// Configures entity mappings for the database context.
    /// </summary>
    /// <param name="modelBuilder">Model builder instance.</param>
    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<EmailLog>(e =>
        {
            e.HasKey(x => x.CorrelationId);

            e.Property(x => x.Status)
                .HasConversion<string>();
        });
    }
}