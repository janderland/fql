defmodule Fql.MixProject do
  use Mix.Project

  def project do
    [
      app: :fql,
      version: "0.1.0",
      elixir: "~> 1.14",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      escript: escript()
    ]
  end

  # Run "mix help compile.app" to learn about applications.
  def application do
    [
      extra_applications: [:logger],
      mod: {Fql.Application, []}
    ]
  end

  # Escript configuration for CLI
  defp escript do
    [main_module: Fql.CLI]
  end

  # Run "mix help deps" to learn about dependencies.
  defp deps do
    [
      # FoundationDB Elixir client
      {:erlfdb, "~> 0.0.5"},
      # CLI argument parsing
      {:optimus, "~> 0.2.0"},
      # Testing
      {:ex_unit, "~> 1.14", only: :test}
    ]
  end
end
