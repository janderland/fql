//! FQL command-line interface
//!
//! Interactive and non-interactive CLI for executing FQL queries.

use clap::{Parser as ClapParser, Subcommand};
use anyhow::Result;

#[derive(ClapParser)]
#[command(name = "fql")]
#[command(about = "FQL - FoundationDB Query Language", long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,

    /// FQL query to execute (non-interactive mode)
    query: Option<String>,
}

#[derive(Subcommand)]
enum Commands {
    /// Run in interactive mode (TUI)
    Interactive,
    
    /// Execute a single query
    Execute {
        /// The FQL query to execute
        query: String,
    },
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();

    match cli.command {
        Some(Commands::Interactive) => {
            run_interactive().await?;
        }
        Some(Commands::Execute { query }) => {
            execute_query(&query).await?;
        }
        None => {
            if let Some(query) = cli.query {
                execute_query(&query).await?;
            } else {
                // Default to interactive mode
                run_interactive().await?;
            }
        }
    }

    Ok(())
}

async fn run_interactive() -> Result<()> {
    println!("FQL Interactive Mode");
    println!("TODO: Implement TUI with ratatui");
    println!("Enter FQL queries (Ctrl+C to exit):");
    
    // TODO: Implement interactive TUI using ratatui
    // Similar to the Go implementation with Bubble Tea
    
    Ok(())
}

async fn execute_query(query_str: &str) -> Result<()> {
    println!("Executing query: {}", query_str);
    
    // Parse the query
    let query = parser::parse(query_str)
        .map_err(|e| anyhow::anyhow!("Parse error: {}", e))?;
    
    // Create engine
    let config = engine::EngineConfig::default();
    let engine = engine::Engine::new(config);
    
    // Execute query
    let results = engine.execute(&query).await
        .map_err(|e| anyhow::anyhow!("Execution error: {}", e))?;
    
    // Display results
    for kv in results {
        println!("{}", parser::format::format(&keyval::Query::KeyValue(kv)));
    }
    
    Ok(())
}
