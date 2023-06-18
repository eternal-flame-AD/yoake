import { styled } from '@mui/material/styles';
import { useState, useEffect, useContext } from 'react';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import { CardContent, CardHeader, Typography, Button, Box } from '@mui/material';
import { useTheme } from '@mui/material/styles';

import { LoginContext } from '../context/LoginContext';
import { postLogout } from '../api/auth';
import { useNavigate } from 'react-router-dom';

const CardItem = styled(Card)(({ theme }) => ({
    ...theme.typography.body2,
    padding: theme.spacing(1),
    textAlign: 'center',
    backgroundColor: "#fff4f5",
}))

function dateInZone(zone?: string) {
    if (zone)
        return new Date(new Date().toLocaleString("en-US", { timeZone: zone }));
    return new Date()
}

function ClockCard() {

    const theme = useTheme();
    const [time, setTime] = useState(new Date());

    useEffect(() => {
        const timer = setInterval(() => {
            setTime(new Date());
        }, 1000);
        return () => clearInterval(timer);
    }, []);

    return (<CardItem>
        <CardHeader title="Clock" sx={{ backgroundColor: theme.palette.primary.main }} />
        <CardContent>
            <Typography variant="h3" component="div">
                {time.toLocaleTimeString()}
            </Typography>
            <hr />
            {
                ["America/Los_Angeles", "America/New_York", "America/Chicago",
                    "Asia/Tokyo", "Asia/Shanghai", "UTC"].map((zone) => {
                        return (
                            <Box key={"clock-" + zone}>
                                <Typography variant="body1" component="div" sx={{ textDecoration: "italic" }}>
                                    {zone}
                                </Typography>
                                <Typography variant="h5" component="div" sx={{ paddingBottom: "1em" }}>
                                    {dateInZone(zone).toLocaleTimeString()}
                                </Typography>
                            </Box>
                        )
                    })
            }
        </CardContent>
    </CardItem>);
}

function HomePage() {

    const theme = useTheme();
    const { auth, refreshAuth } = useContext(LoginContext);
    const navigate = useNavigate();

    return (

        <Grid container spacing={4} sx={{ minWidth: "80%", flexGrow: 2 }} columns={{ xs: 4, sm: 8, md: 12 }}>
            <Grid item xs={4} sm={8} md={8}>
                <CardItem>
                    <CardHeader title="Welcome" sx={{ backgroundColor: theme.palette.primary.main }} />
                    <CardContent>
                        <Typography variant="h6" component="div">
                            夜明け前が一番暗い。
                        </Typography>
                        <Typography variant="body2" component="div">
                            The darkest hour is just before the dawn.
                        </Typography>
                        <hr />
                        <Typography variant="body2" component="div">
                            This is yoake.yumechi.jp, Yumechi's <abbr title="Personal Information Manager" className="initialism">PIM</abbr>. <br />
                            Built with axum and React.
                        </Typography>
                        <hr />
                        {
                            auth.valid ?
                                <>
                                    <Typography variant="body2" component="div">
                                        Welcome, {auth.display_name}
                                    </Typography>
                                    <Button variant="contained" onClick={() => {
                                        postLogout().then(refreshAuth).catch(console.error);
                                    }}>Logout</Button>
                                </>
                                :
                                <>
                                    <Typography variant="body2" component="div">
                                        You are not logged in
                                    </Typography>
                                    <Button variant="contained" onClick={() => navigate("/login")}>Login</Button>
                                </>
                        }

                    </CardContent>
                </CardItem>
            </Grid>
            <Grid item xs={4} sm={8} md={4}>
                <ClockCard />
            </Grid>
        </Grid>
    );
}

export default HomePage;