import { createContext } from "react";
import { AuthInfo, EmptyAuthInfo } from "../api/auth";

export const LoginContext = createContext({
    auth: EmptyAuthInfo,
    setAuth: (_auth: AuthInfo) => { },
    refreshAuth: () => { },
});