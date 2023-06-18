import { useEffect, useState } from "react";
import { Grading, getGradings, GetGradingsResponse } from "../api/canvas_lms";
import { Typography, Box, Button, Link } from "@mui/material";
import { DataGrid, GridCellParams, GridColDef } from "@mui/x-data-grid";
import TimeBody from "../components/TimeBody";


const renderTimeCell = (params: GridCellParams) => {
    const value = params.value as Date | null;
    return <TimeBody time={value} />;
}


const grading_columns: GridColDef[] = [
    {
        field: "course_code",
        headerName: "Course",
        width: 100,
    },
    {
        field: "name",
        headerName: "Name",
        width: 200,
        renderCell: (params) => {
            const row = params.row as Grading;
            return <Link href={row.assignment_url}
                target="_blank"
                rel="noreferrer noopener"
                color="inherit"
            >{row.name}</Link>;
        }
    },
    {
        field: "grade",
        headerName: "Grade",
        minWidth: 250,
        valueGetter: (params) => {
            const row = params.row as Grading;
            if (row.grade_hidden) {
                return "Hidden";
            } else if (row.score === null) {
                return "Not Graded";
            }
            const percentage = row.score! / row.possible_points;
            return `${row.score!.toFixed(2)} (${row.grade}) / ${row.possible_points} (${(percentage * 100).toFixed(2)}%)`;
        },
        flex: 1,
        cellClassName: (params) => {
            const row = params.row as Grading;
            if (row.grade_hidden) {
                return "grade_hidden";
            } else if (row.score === null) {
                return "grade_not_graded";
            }
            const percentage = row.score! / row.possible_points;
            if (percentage < 0.6) {
                return "grade_bad";
            } else if (percentage < 0.8) {
                return "grade_ok";
            } else if (percentage < 1.0) {
                return "grade_good";
            } else {
                return "grade_perfect";
            }
        }
    },
    {
        field: "graded_at",
        headerName: "Graded At",
        minWidth: 100,
        valueGetter: (params) => {
            const row = params.row as Grading;
            return row.graded_at ? new Date(row.graded_at) : null;
        },
        flex: 1,
        renderCell: renderTimeCell
    },
    {
        field: "posted_at",
        headerName: "Posted At",
        minWidth: 100,
        valueGetter: (params) => {
            const row = params.row as Grading;
            return row.posted_at ? new Date(row.posted_at) : null;
        },
        flex: 1,
        renderCell: renderTimeCell
    },
    {
        field: "updated_at",
        headerName: "Updated At",
        minWidth: 100,
        valueGetter: (params) => {
            const row = params.row as Grading;
            let ret = null;
            if (row.posted_at !== null)
                ret = new Date(row.posted_at);
            if (row.graded_at !== null && (ret == null || new Date(row.graded_at) > ret))
                ret = new Date(row.graded_at);
            return ret;
        },
        flex: 1,
        renderCell: renderTimeCell
    }
];

function GradePage() {
    const [mounted, setMounted] = useState<boolean>(false);
    const [gradings, setGradings] = useState<GetGradingsResponse | null>(null);
    const [updating, setUpdating] = useState<boolean>(false);


    const updateGradings = (force: boolean) => {
        setUpdating(true);
        getGradings(force).then((gradings) => {
            setGradings(gradings);
        }).catch((error) => {
            console.error(error);
        }).finally(() => {
            setUpdating(false);
        });
    }

    if (!mounted) {
        setMounted(true);
        updateGradings(false);
    }

    useEffect(() => {
        const interval = setInterval(() => updateGradings(false), 5 * 60 * 1000);
        return () => clearInterval(interval);
    }, []);


    return (
        <>
            <Typography variant="h4" component="h1" gutterBottom>
                Grades
                {
                    gradings ?
                        <Box>
                            <Typography variant="caption" display="block" gutterBottom>
                                Last updated: <TimeBody time={new Date(gradings.last_updated)} />
                            </Typography>
                            <Button
                                variant="contained" onClick={() => updateGradings(true)} disabled={updating}
                                sx={{ marginBottom: 2 }}
                            >Refresh</Button>
                            <DataGrid
                                rows={gradings.response}
                                columns={grading_columns}
                                density="compact"
                                getRowId={(row) => row.submission_id}
                                autoHeight={true}
                                sortModel={[{
                                    field: "updated_at",
                                    sort: "desc",
                                }]}
                                initialState={{
                                    pagination: {
                                        paginationModel: {
                                            pageSize: 25,
                                        }
                                    },
                                }}
                                sx={{
                                    "& .grade_hidden": {
                                        color: "grey",
                                    },
                                    "& .grade_not_graded": {
                                        color: "grey",
                                    },
                                    "& .grade_bad": {
                                        color: "red",
                                    },
                                    "& .grade_ok": {
                                        color: "orange",
                                    },
                                    "& .grade_good": {
                                        color: "blue",
                                    },
                                    "& .grade_perfect": {
                                        color: "green",
                                    },
                                }}
                            />
                        </Box>
                        : "Loading..."
                }
            </Typography>
        </>
    )
}

export default GradePage