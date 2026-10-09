--!strict
-- Tells the server this client has loaded: ClientReady in
-- design/Game/Networking.md.

local ReplicatedStorage = game:GetService("ReplicatedStorage")

local Remotes = require(ReplicatedStorage.Shared.Remotes)

local ReadyController = {}

function ReadyController:start(): ()
	Remotes.event("ClientReady"):FireServer()
end

return ReadyController
