import { Tooltip, Typography } from "@mui/material";
import { useEffect, useState } from "react";

interface TimeBodyProps {
    time: Date | null;
}


function formatRelativeTime(time: Date): [string, number] {
    const now = new Date();
    const diff = now.getTime() - time.getTime();
    if (diff >= 0) {
        if (diff < 1000) {
            return ["just now", 1000];
        }
        if (diff < 60 * 1000) {
            return [`${Math.floor(diff / 1000)} seconds ago`, 1000];
        } else if (diff < 60 * 60 * 1000) {
            return [`${Math.floor(diff / 1000 / 60)} minutes ago`, 60 * 1000];
        } else if (diff < 24 * 60 * 60 * 1000) {
            return [`${Math.floor(diff / 1000 / 60 / 60)} hours ago`, 60 * 60 * 1000];
        } else {
            return [`${Math.floor(diff / 1000 / 60 / 60 / 24)} days ago`, 24 * 60 * 60 * 1000];
        }
    } else {
        if (diff > -1000) {
            return ["just now", 1000];
        }
        if (diff > -60 * 1000) {
            return [`${Math.floor(-diff / 1000)} seconds from now`, 1000];
        } else if (diff > -60 * 60 * 1000) {
            return [`${Math.floor(-diff / 1000 / 60)} minutes from now`, 60 * 1000];
        } else if (diff > -24 * 60 * 60 * 1000) {
            return [`${Math.floor(-diff / 1000 / 60 / 60)} hours from now`, 60 * 60 * 1000];
        } else {
            return [`${Math.floor(-diff / 1000 / 60 / 60 / 24)} days from now`, 24 * 60 * 60 * 1000];
        }
    }
}

export default function TimeBody(props: TimeBodyProps) {
    const { time } = props;
    if (time === null) {
        return <>N/A</>;
    }

    const [relativeTime, setRelativeTime] = useState(formatRelativeTime(time)[0]);

    useEffect(() => {
        let timer: number | null = null;
        const update = () => {
            const [relativeTime, interval] = formatRelativeTime(time);
            setRelativeTime(relativeTime);
            timer = setTimeout(update, interval);
        }
        update();
        return () => clearTimeout(timer!);
    }, [time]);


    return <Tooltip title={time.toLocaleString()} sx={{ display: "inline-block" }}>
        <Typography variant="body2" color="textSecondary"> {relativeTime}</Typography>
    </Tooltip>

}