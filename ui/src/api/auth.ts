import { makeJSONRequest } from "./request";

export type Role = "Admin" | "User";

export interface AuthInfo {
    valid: boolean,
    user: string,
    display_name: string,
    roles: Role[],
}

export const EmptyAuthInfo: AuthInfo = {
    valid: false,
    user: "",
    display_name: "",
    roles: [],
}

export async function getLoginState(): Promise<AuthInfo> {
    const response = await makeJSONRequest<AuthInfo>("/api/auth/self", "GET");
    return response.data;
}

interface PostLoginParams {
    username: string,
    password: string,
}

export async function postLogin(params: PostLoginParams): Promise<AuthInfo> {
    const response = await makeJSONRequest<AuthInfo>("/api/auth/login", "POST", {
        username: params.username,
        password: params.password,
    });
    if (response.status != "Ok") {
        throw new Error(response.message);
    }
    return response.data;
}

export async function postLogout(): Promise<void> {
    await makeJSONRequest<void>("/api/auth/logout", "POST");
}