import { useEffect } from "react";


type FramePageProps = {
    url: string;
    sx?: React.CSSProperties;
}

export default function FramePage(props: FramePageProps) {
    useEffect(() => {
        console.log("FramePage mounted, url: " + props.url);
        return () => {
            console.log("FramePage unmounted, url: " + props.url);
        }
    }, [props.url]);
    return (
        <iframe src={props.url} style={{
            width: "100%", height: "100%", border: "none", aspectRatio: "16/9", ...props.sx
        }}></iframe>
    )
}