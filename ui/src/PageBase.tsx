import { ReactNode, useContext, useState } from 'react'
import './App.css'
import { AppBar, Button, Divider, IconButton, Toolbar, Drawer, List, ListItem, ListItemButton, ListItemText } from '@mui/material'
import Typography from '@mui/material/Typography'
import { Box, Container } from '@mui/system'
import MenuIcon from '@mui/icons-material/Menu'
import HomeIcon from '@mui/icons-material/Home'
import GradeIcon from '@mui/icons-material/Grade'
import CampaignIcon from '@mui/icons-material/Campaign'
import MedicationIcon from '@mui/icons-material/Medication'
import { useMatches, useNavigate } from 'react-router-dom'
import { LoginContext } from './context/LoginContext'
import { ServerInfo, getServerInfo } from './api/server_info'

interface PersistentPage {
    path: string;
    element: ReactNode;
};

const drawerWidth = 240;

function routerMatch(matches: { pathname: string }[], path: string) {
    for (let i = 0; i < matches.length; i++) {
        if (matches[i].pathname === path) {
            return true;
        }
    }
    return false;
}


function PageBase({ children, persistentPages }: { children?: ReactNode, persistentPages: PersistentPage[] }) {
    const navigate = useNavigate();
    const matches = useMatches();

    const { auth } = useContext(LoginContext);
    const [openMenu, setOpenMenu] = useState(false);

    const handleMenuToggle = () => {
        setOpenMenu(!openMenu);
    };

    const [server_info, setServerInfo] = useState<ServerInfo | null>(null);
    if (!server_info) {
        getServerInfo().then((serverInfo) => {
            setServerInfo(serverInfo);
        }).catch((error) => {
            console.error(error);
        });
    }

    const drawer = (
        <div>
            <Toolbar />
            <Divider />
            <List>
                {[
                    { key: "home", name: "Home", icon: <HomeIcon />, path: "/" },
                    { key: "grades", name: "Grades", icon: <GradeIcon />, path: "/grades" },
                    { key: "meds", name: "Meds", icon: <MedicationIcon />, path: "/meds" },
                    { key: "gotify", name: "Gotify", icon: <CampaignIcon />, path: "/gotify_ui" },
                ].map((item) => (
                    <ListItem key={item.key}
                        onClick={() => {
                            navigate(item.path);
                        }}
                        sx={{ color: routerMatch(matches, item.path) ? "#f00" : "#000" }}
                    >
                        <ListItemButton>
                            {item.icon}
                            <ListItemText primary={item.name} />
                        </ListItemButton>
                    </ListItem>
                ))
                }
            </List>
        </div>
    )

    return (
        <Box sx={{ display: "flex" }}>

            <AppBar
                position="fixed"
                sx={{
                    width: { md: `calc(100% - ${drawerWidth}px)` },
                    ml: { md: `${drawerWidth}px` },
                }}
            >
                <Container maxWidth="lg">
                    <Toolbar disableGutters>
                        <IconButton edge="start" color="inherit" aria-label="menu"
                            onClick={handleMenuToggle}
                            sx={{ mr: 2, display: { md: 'none' } }}
                        >
                            <MenuIcon />
                        </IconButton>
                        <Typography variant="h6" component="a"
                            onClick={() => { navigate("/") }}
                            sx={{ color: 'inherit', textDecoration: 'none', display: { md: 'flex' }, }} >
                            夜明け
                        </Typography>
                        <Box sx={{ flexGrow: 0 }}>
                            <Typography variant="body1" component="i"
                                sx={{ color: '#808080', display: 'block', marginLeft: '0.6em' }} >
                                {server_info?.version ? `${server_info.version} (${server_info.profile})` : "Unknown"}
                            </Typography>
                        </Box>
                        <Box sx={{ flexGrow: 1 }} />
                        {
                            !auth.valid ?
                                <Button color="inherit" onClick={() => navigate("/login")}>Login</Button>
                                :
                                <Button disabled color="inherit">{auth.display_name}</Button>
                        }
                    </Toolbar>
                </Container>
            </AppBar >
            <Box
                component="nav"
                sx={{ width: { md: drawerWidth }, flexShrink: { md: 0 } }}
            >
                <Drawer
                    container={window.document.body}
                    variant="temporary"
                    open={openMenu}
                    onClose={handleMenuToggle}
                    ModalProps={{
                        keepMounted: true, // Better open performance on mobile.
                    }}
                    sx={{
                        display: { xs: 'block', md: 'none' },
                        '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
                    }}
                >
                    {drawer}
                </Drawer>
                <Drawer
                    variant="permanent"
                    sx={{
                        display: { xs: 'none', md: 'block' },
                        '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
                    }}
                    open
                >
                    {drawer}
                </Drawer>
            </Box>

            {
                (children ?
                    <Box
                        sx={{ width: "100%", display: "block", marginTop: "4em", padding: "1em" }}>
                        {children}
                    </Box> : null)
            }

            {
                persistentPages.map((page) => {
                    return (
                        <Box key={page.path} sx={{
                            width: "100%", display: routerMatch(matches, page.path)
                                ? "block" : "none"
                            , marginTop: "4em", padding: "1em"
                        }}>
                            {page.element}
                        </Box>
                    )
                })
            }
        </Box >
    )
}

export default PageBase
