defmodule Fql.Application do
  @moduledoc """
  The FQL Application.

  This module defines the supervision tree for the FQL application.
  """

  use Application

  @impl true
  def start(_type, _args) do
    children = [
      # Add children here as needed
      # {Fql.Worker, arg}
    ]

    opts = [strategy: :one_for_one, name: Fql.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
