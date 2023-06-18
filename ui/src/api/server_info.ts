import { makeJSONRequest } from "./request";

export interface ServerInfo {
    version: string,
    profile: string,
};

export async function getServerInfo(): Promise<ServerInfo> {
    const response = await makeJSONRequest<ServerInfo>("/api/server_info", "GET");
    return response.data;
}