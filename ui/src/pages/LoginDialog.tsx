import {
    Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle,
    TextField, Button, Alert
} from "@mui/material";
import { Box } from "@mui/system";
import { postLogin } from "../api/auth";
import { LoginContext } from "../context/LoginContext";
import { useContext, useRef, useState } from "react";
import { Navigate, NavigateFunction, useNavigate } from "react-router-dom";


type Props = {
    open: boolean;
    onClose?: (success: boolean, navigate: NavigateFunction) => void;
}
export default function LoginDialog({ open, onClose }: Props) {

    const { auth, refreshAuth } = useContext(LoginContext);
    const usernameNode = useRef<HTMLInputElement>(null);
    const passwordNode = useRef<HTMLInputElement>(null);
    const [loginError, setLoginError] = useState<string | null>(null);
    const navigate = useNavigate();

    return auth.valid ? <Navigate to="/" /> :

        (
            <Dialog open={open} onClose={() => onClose?.(false, navigate)}>
                <Box component="form" onSubmit={(e) => {
                    e.preventDefault();
                    if (
                        !usernameNode.current ||
                        !passwordNode.current ||
                        !usernameNode.current
                    ) {
                        setLoginError("Internal error: missing form elements.");
                        return;
                    }
                    const username = usernameNode.current?.value;
                    const password = passwordNode.current?.value;
                    postLogin({ username: username!, password: password! }).then((auth) => {
                        console.log("Got new auth state: ", auth);
                        refreshAuth();
                        onClose?.(true, navigate);
                    }).catch((error) => {
                        console.error(error);
                        setLoginError(error.message);
                        refreshAuth();
                    });
                }}>
                    <DialogTitle>Login</DialogTitle>
                    <DialogContent>
                        {
                            loginError ?
                                <Alert severity="error">{loginError}</Alert>
                                : <Alert severity="info">Present Credentials.</Alert>
                        }
                        <DialogContentText>
                            Userpass login:
                        </DialogContentText>

                        <TextField
                            id="username"
                            label="Username"
                            variant="standard"
                            fullWidth
                            inputRef={usernameNode}
                        />
                        <TextField
                            id="password"
                            label="Password"
                            variant="standard"
                            type="password"
                            fullWidth
                            inputRef={passwordNode}
                        />
                    </DialogContent>
                    <DialogActions>
                        {
                            onClose ?
                                <Button onClick={() => onClose?.(false, navigate)}>Back</Button>
                                :
                                <Button onClick={() => navigate("/")}>Cancel</Button>
                        }
                        <Button type="submit">Login</Button>
                    </DialogActions>
                </Box>
            </Dialog >
        )
}