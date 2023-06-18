import { useState } from "react";
import { Accordion, AccordionDetails, AccordionSummary, Typography, TextField, Box, Divider, Button } from "@mui/material"
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { parseShorthand, Medication, postDirective, patchDirective, getDirectives, deleteDirective, formatShorthand } from "../api/med_directive";
import { MedicationLog, deleteMedicationLog, getMedicationLog, postMedicationLog, projectNextDose } from "../api/med_log";
import { DataGrid, GridColDef } from "@mui/x-data-grid";
import TimeBody from "../components/TimeBody";
import { format_rust_naive_date, parse_rust_naive_date } from "../api/time";


function DirectionEditor({ onUpdate }: { onUpdate?: (medication: Medication) => void }) {
    const [direction, setDirection] = useState<Medication | null>(null);

    return (
        <Accordion>
            <AccordionSummary
                expandIcon={<ExpandMoreIcon />}
            >
                <Typography>Edit Direction</Typography>
            </AccordionSummary>
            <AccordionDetails>
                <TextField fullWidth
                    label="Shorthand" variant="outlined"
                    onChange={(event) => {
                        parseShorthand(event.target.value)
                            .then((medication) => {
                                medication.uuid = direction?.uuid || "";
                                setDirection(medication);
                            })
                    }}
                />
                <TextField fullWidth
                    label="UUID" variant="outlined"
                    value={direction?.uuid || ""} onChange={(event) => {
                        setDirection({ ...direction!, uuid: event.target.value });
                    }} />
                <TextField label="Name"
                    variant="standard"
                    value={direction?.name || ""} onChange={(event) => {
                        setDirection({ ...direction!, name: event.target.value });
                    }} />
                <TextField label="Dosage"
                    type="number"
                    variant="standard"
                    value={direction?.dosage || ""} onChange={(event) => {
                        setDirection({ ...direction!, dosage: parseInt(event.target.value) });
                    }} />
                <TextField label="Dosage Unit"
                    variant="standard"
                    value={direction?.dosage_unit || ""} onChange={(event) => {
                        setDirection({ ...direction!, dosage_unit: event.target.value });
                    }} />
                <br />
                <TextField label="Period Hours"
                    type="number"
                    variant="standard"
                    value={direction?.period_hours || ""} onChange={(event) => {
                        setDirection({ ...direction!, period_hours: parseInt(event.target.value) });
                    }} />
                <TextField label="Flags"
                    variant="standard"
                    value={direction?.flags || ""} onChange={(event) => {
                        setDirection({ ...direction!, flags: event.target.value });
                    }} />
                <TextField label="Options"
                    variant="standard"
                    value={direction?.options || ""} onChange={(event) => {
                        setDirection({ ...direction!, options: event.target.value });
                    }} />
                <Divider sx={{ margin: 1 }} />
                <Button variant="contained" onClick={() => {
                    ((direction?.uuid) ? patchDirective : postDirective)
                        (direction!).then((response) => {
                            setDirection(response);
                            onUpdate?.(response);
                        });
                }}>{direction?.uuid ? "Update" : "Create"}</Button>
                <Button variant="contained"
                    disabled={!direction?.uuid}
                    onClick={() => {
                        deleteDirective(direction!.uuid).then(() => {
                            let deleted = direction!;
                            setDirection(null);
                            onUpdate?.(deleted);
                        });
                    }}>Delete</Button>
            </AccordionDetails>
        </Accordion >
    )
}


interface MedPanelForm {
    dosage: number;
    time_actual: Date | null;
}

function MedPanel({ medication }: { medication: Medication }) {
    const [shorthand, setShorthand] = useState("");
    const [mounted, setMounted] = useState(false);
    const [log, setLog] = useState<MedicationLog[]>([]);
    const [nextDose, setNextDose] = useState<MedicationLog | null>(null);
    const [form, setForm] = useState<MedPanelForm>({ dosage: medication.dosage, time_actual: null });

    const updateShorthand = () => {
        formatShorthand(medication).then((shorthand) => {
            setShorthand(shorthand);
        });
    }


    const updateLog = () => {
        getMedicationLog(medication.uuid, { limit: 100 }).then((log) => {
            setLog(log);
        });
    }

    const updateNextDose = () => {
        projectNextDose(medication.uuid).then((nextDose) => {
            setNextDose(nextDose);
        });
    }


    if (!mounted) {
        updateShorthand();
        updateLog();
        updateNextDose();
        setMounted(true);
    }

    const med_log_columns: GridColDef[] = [
        {
            field: "Action",
            headerName: "Action", minWidth: 100,
            renderCell: (params) => {
                const log = params.row as MedicationLog;
                return (
                    <Button variant="contained" onClick={() => {
                        if (confirm(`Delete log entry ${log.uuid}?`)) {
                            deleteMedicationLog(log.med_uuid, log.uuid).then(() => {
                                updateLog();
                            });
                        }
                    }}>Delete</Button>
                )
            }
        },
        { field: 'dosage', headerName: 'Dosage', minWidth: 100 },
        {
            field: 'time_actual', headerName: 'Time Actual', minWidth: 200,
            renderCell: (params) =>
                <TimeBody time={parse_rust_naive_date(params.value as string)} />

        },
        {
            field: 'time_expected', headerName: 'Time Expected', minWidth: 200,
            renderCell: (params) =>
                <TimeBody time={parse_rust_naive_date(params.value as string)} />
        },
        {
            field: 'dose_offset', headerName: 'Dose Offset', minWidth: 100,
            renderCell: (params) => {
                const log = params.row as MedicationLog;
                return (
                    <>
                        {
                            log.dose_offset.toFixed(2)
                        }
                    </>
                )
            }
        },
    ];


    return (
        <Accordion>
            <AccordionSummary
                expandIcon={<ExpandMoreIcon />}
            >
                <Typography>
                    {shorthand}
                    {
                        nextDose ? <Typography variant="caption" component="span"> Next Dose:
                            {
                                nextDose.dose_offset >= 0 ? <span style={{ color: "red" }}> now </span> :
                                    <TimeBody time={
                                        parse_rust_naive_date(nextDose.time_expected)
                                    } />}
                        </Typography> : null
                    }
                </Typography>
            </AccordionSummary>
            <AccordionDetails>
                <Box sx={{ padding: "1em", textAlign: "left" }}>
                    UUID: {medication.uuid}
                </Box>
                <Box sx={{ padding: "1em", textAlign: "left" }}>
                    <TextField label="Dose" variant="standard" type="number" value={form.dosage} onChange={(event) => {
                        setForm({ ...form, dosage: parseInt(event.target.value) });
                    }} />
                    <TextField label="Time" variant="standard" type="datetime-local"
                        onChange={(event) => {
                            setForm({ ...form, time_actual: new Date(event.target.value) });
                        }} />
                    <Button fullWidth variant="contained" onClick={() => {
                        const content = {
                            ...nextDose!,
                            dosage: form.dosage,
                            time_actual: format_rust_naive_date(form.time_actual || new Date()),
                            med_uuid: medication.uuid,
                        }
                        postMedicationLog(content).then(() => {
                            setForm({ dosage: medication.dosage, time_actual: null });
                            updateLog();
                            updateNextDose();
                        });
                    }}>Log</Button>
                </Box>
                <DataGrid
                    rows={log}
                    density="compact"
                    columns={med_log_columns}
                    getRowId={(row) => row.uuid}
                    autoHeight
                    initialState={{
                        pagination: {
                            paginationModel: {
                                pageSize: 25,
                            }
                        },
                    }}
                />
            </AccordionDetails>
        </Accordion >
    )
}

function MedsPage() {
    const [medications, setMedications] = useState<Medication[] | null>(null);
    const [mounted, setMounted] = useState(false);

    const refreshMedications = () => {
        getDirectives().then((response) => {
            setMedications(response);
        });
    }

    if (!mounted) {
        refreshMedications();
        setMounted(true);
    }

    return (
        <Box
            sx={{
                '& .MuiTextField-root': { margin: 1 },
            }}
        >
            {
                medications?.map((medication) => {
                    return <MedPanel medication={medication} key={medication.uuid} />
                })
            }
            <DirectionEditor onUpdate={() => {
                refreshMedications();
            }} />
        </Box>
    )
}

export default MedsPage