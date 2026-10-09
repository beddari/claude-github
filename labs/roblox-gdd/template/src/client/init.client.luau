--!strict
-- The one client LocalScript. It starts every controller in Controllers/ in
-- name order.

local ReplicatedStorage = game:GetService("ReplicatedStorage")

local Types = require(ReplicatedStorage.Shared.Types)

local controllersFolder = script:WaitForChild("Controllers")
local modules = controllersFolder:GetChildren()
table.sort(modules, function(a: Instance, b: Instance): boolean
	return a.Name < b.Name
end)

for _, module in modules do
	if module:IsA("ModuleScript") then
		local controller = (require :: any)(module) :: Types.Controller
		controller:start()
	end
end
