--!strict
-- Logs each player whose client has loaded. The first remote of the game,
-- and what the setup's playtest looks for: design/Game/Networking.md.

local Players = game:GetService("Players")
local ReplicatedStorage = game:GetService("ReplicatedStorage")

local Config = require(ReplicatedStorage.Shared.Config.Game)
local Remotes = require(ReplicatedStorage.Shared.Remotes)

local lastReady: { [Player]: number } = {}

local ReadyService = {}

function ReadyService:start(): ()
	Remotes.event("ClientReady").OnServerEvent:Connect(function(player: Player, ...: unknown)
		-- Rate first, so a flooding client cannot flood the log either.
		local now = os.clock()
		local last = lastReady[player]
		if last and now - last < Config.CLIENT_READY_COOLDOWN then
			return
		end
		lastReady[player] = now
		-- ClientReady takes no arguments; anything else is a modified client.
		if select("#", ...) ~= 0 then
			warn(`[Server] ClientReady from {player.Name} with arguments; ignored`)
			return
		end
		print(`[Server] ClientReady from {player.Name}`)
	end)

	Players.PlayerRemoving:Connect(function(player: Player)
		lastReady[player] = nil
	end)
end

return ReadyService
