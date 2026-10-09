--!strict
-- Tuning values. The GDD quotes these; a change here updates the page that
-- describes the mechanic in the same task.

local Game = {
	-- Seconds a player must wait before ClientReady counts again.
	CLIENT_READY_COOLDOWN = 5,
}

return table.freeze(Game)
