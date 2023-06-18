import { useContext } from "react";
import { LoginContext } from "../context/LoginContext";
import { Role } from "../api/auth";
import LoginDialog from "../pages/LoginDialog";

export function EnsureRole({ children, role }: { children: React.ReactNode, role: Role }) {
    const { auth } = useContext(LoginContext);

    if (auth.roles.includes(role)) {
        return <>{children}</>;
    } else {
        return <><LoginDialog open={true} onClose={(success) => {
            if (!success) {
                window.history.back();
            }
        }} /></>;
    }
}