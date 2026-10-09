--!strict
-- Every RemoteEvent and RemoteFunction in the game, one name per line.
-- Each name has a "### Name" section in design/Game/Networking.md, and
-- gdd-code-sync's check-gdd.sh fails when the two lists differ.

local ReplicatedStorage = game:GetService("ReplicatedStorage")
local RunService = game:GetService("RunService")

local REMOTE_EVENTS: { string } = {
	"ClientReady",
}

local REMOTE_FUNCTIONS: { string } = {}

local FOLDER_NAME = "Remotes"

-- The server creates the instances; clients wait for them to replicate.
local folder: Instance
if RunService:IsServer() then
	local created = Instance.new("Folder")
	created.Name = FOLDER_NAME
	for _, name in REMOTE_EVENTS do
		local remote = Instance.new("RemoteEvent")
		remote.Name = name
		remote.Parent = created
	end
	for _, name in REMOTE_FUNCTIONS do
		local remote = Instance.new("RemoteFunction")
		remote.Name = name
		remote.Parent = created
	end
	created.Parent = ReplicatedStorage
	folder = created
else
	folder = ReplicatedStorage:WaitForChild(FOLDER_NAME)
end

local Remotes = {}

function Remotes.event(name: string): RemoteEvent
	assert(table.find(REMOTE_EVENTS, name), `Unknown RemoteEvent {name}; declare it in Remotes.luau`)
	local remote = folder:WaitForChild(name)
	assert(remote:IsA("RemoteEvent"), `{name} is not a RemoteEvent`)
	return remote
end

function Remotes.func(name: string): RemoteFunction
	assert(table.find(REMOTE_FUNCTIONS, name), `Unknown RemoteFunction {name}; declare it in Remotes.luau`)
	local remote = folder:WaitForChild(name)
	assert(remote:IsA("RemoteFunction"), `{name} is not a RemoteFunction`)
	return remote
end

return Remotes
