--!strict
-- The one server Script. It creates the remotes, then starts every service
-- in Services/ in name order.

local ServerScriptService = game:GetService("ServerScriptService")
local ReplicatedStorage = game:GetService("ReplicatedStorage")

local Types = require(ReplicatedStorage.Shared.Types)
require(ReplicatedStorage.Shared.Remotes)

local servicesFolder = ServerScriptService.Server.Services
local modules = servicesFolder:GetChildren()
table.sort(modules, function(a: Instance, b: Instance): boolean
	return a.Name < b.Name
end)

for _, module in modules do
	if module:IsA("ModuleScript") then
		local service = (require :: any)(module) :: Types.Service
		service:start()
	end
end

print("[Server] started")
