--!strict
-- Types shared by the server and the client. Server-only types stay in
-- src/server, because everything in src/shared replicates to every client.

export type Service = {
	start: (self: Service) -> (),
}

export type Controller = {
	start: (self: Controller) -> (),
}

return {}
