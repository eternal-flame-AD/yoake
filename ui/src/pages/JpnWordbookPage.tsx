import { Accordion, AccordionDetails, AccordionSummary, Alert, Button, Container, List, ListItem, ListItemText, Paper, TextField, Typography } from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { LookupResult, WordbookItem, comboSearchWord, comboSearchWordTop, downloadWordbookCsv, getWordbook, storeWordbook } from "../api/jpn_wordbook";
import { DataGrid } from "@mui/x-data-grid";
import { LoginContext } from "../context/LoginContext";

export default function JpnWordbookPage() {
    const [keyword, setKeyword] = useState<string>("");
    const [lookupResult, setLookupResult] = useState<LookupResult[] | null>(null);
    const [lookupError, setLookupError] = useState<string | null>(null);
    const [lookupStoreCount, setLookupStoreCount] = useState<number>(0);
    const [wordbookRows, setWordbookRows] = useState<WordbookItem[] | null>(null);
    const [wordbookPaginationModel, setWordbookPaginationModel] = useState({
        page: 0,
        pageSize: 100,
    })
    const [wordbookError, setWordbookError] = useState<string | null>(null);
    const wordbookQueryOptions = useMemo(() => {
        return {
            until: (wordbookRows && wordbookRows.length > 0) ? wordbookRows[wordbookRows.length - 1]?.created : undefined,
            limit: wordbookPaginationModel.pageSize,
        }
    }, [wordbookPaginationModel]);
    useEffect(() => {
        getWordbook(wordbookQueryOptions).then((result) => {
            let rows = wordbookRows ?? [];
            for (const new_row of result) {
                if (rows.find((row) => row.uuid === new_row.uuid)) {
                    continue;
                }
                rows.push(new_row);
            }
            setWordbookRows(rows);
        }).catch((error) => {
            setWordbookError(error.message);
        });
    }, [wordbookQueryOptions, lookupStoreCount]);
    return (
        <LoginContext.Consumer>
            {
                ({ auth }) => (

                    <Container>
                        <Paper sx={{ padding: "1em" }}>
                            <Typography variant="h4" component="h1" gutterBottom>
                                Lookup word
                            </Typography>
                            {
                                lookupError ?
                                    <Alert severity="error">{lookupError}</Alert>
                                    : null
                            }
                            <TextField label="Word" variant="outlined" value={keyword} onChange={(event) => {
                                setKeyword(event.target.value);
                            }} />
                            <Button variant="contained" onClick={() => {
                                comboSearchWord(keyword).then((result) => {
                                    setLookupError(null);
                                    setLookupResult(result);
                                }).catch((error) => {
                                    setLookupError(error.message);
                                });
                            }}>Search</Button>
                            <Button variant="contained" onClick={() => {
                                comboSearchWordTop(keyword).then((result) => {
                                    setLookupResult([result]);
                                }).catch((error) => {
                                    setLookupError(error.message);
                                });
                            }}>Top</Button>
                            {
                                lookupResult && lookupResult.map((result) => {
                                    return (
                                        <Accordion key={result.ja}>
                                            <AccordionSummary>
                                                {result.ja}
                                            </AccordionSummary>
                                            <AccordionDetails>
                                                {
                                                    auth.roles.includes("Admin") ?
                                                        <Button variant="contained" onClick={() => {
                                                            storeWordbook(result)
                                                                .then(() => {
                                                                    setLookupStoreCount(lookupStoreCount + 1);
                                                                })
                                                                .catch((error) => {
                                                                    setWordbookError(error.message);
                                                                });
                                                        }}>Store</Button>
                                                        : null
                                                }
                                                <Typography variant="h6" component="div" gutterBottom>
                                                    解説
                                                </Typography>
                                                <List dense={true}>
                                                    {
                                                        result.jm?.map((item) => {
                                                            return (
                                                                <ListItem key={item}>
                                                                    -&nbsp;<ListItemText primary={item} />
                                                                </ListItem>
                                                            )
                                                        })
                                                    }
                                                </List>
                                                <Typography variant="h6" component="div" gutterBottom>
                                                    英訳
                                                </Typography>
                                                <List dense={true}>
                                                    {
                                                        result.en?.map((item) => {
                                                            return (
                                                                <ListItem key={item}>
                                                                    -&nbsp;<ListItemText primary={item} />
                                                                </ListItem>
                                                            )
                                                        })
                                                    }
                                                </List>
                                                <Typography variant="h6" component="div" gutterBottom>
                                                    例文
                                                </Typography>
                                                <List dense={true}>
                                                    {
                                                        result.ex?.map((item) => {
                                                            return (
                                                                <ListItem key={item}>
                                                                    -&nbsp;<ListItemText primary={item} />
                                                                </ListItem>
                                                            )
                                                        })
                                                    }
                                                </List>
                                            </AccordionDetails>
                                        </Accordion>
                                    )
                                })
                            }
                        </Paper>
                        <Paper sx={{ padding: "1em", marginTop: "1em" }}>
                            <Typography variant="h4" component="h1" gutterBottom>
                                Wordbook
                            </Typography>
                            {
                                wordbookError ?
                                    <Alert severity="error">{wordbookError}</Alert>
                                    : null
                            }
                            <Button variant="contained" onClick={() => {
                                downloadWordbookCsv(true);
                            }}>Download</Button>
                            <Button variant="contained" onClick={() => {
                                downloadWordbookCsv(false);
                            }}>Download (no header)</Button>
                            <DataGrid
                                paginationMode="client"
                                rows={wordbookRows ?? []}
                                columns={
                                    [
                                        { field: "ja", headerName: "Japanese", width: 200 },
                                        { field: "fu", headerName: "Furigana", width: 200 },
                                        { field: "en", headerName: "English", width: 200 },
                                        { field: "ex", headerName: "Example", width: 200 },
                                        { field: "jm", headerName: "Meaning", width: 200 },
                                    ]
                                }
                                getRowId={(row) => row.uuid}
                                onPaginationModelChange={(model) => {
                                    console.log("Pagination model changed: ", model);
                                    setWordbookPaginationModel(model);
                                }}
                                initialState={
                                    {
                                        pagination: {
                                            paginationModel: {
                                                pageSize: 100,
                                            }
                                        }
                                    }
                                }
                            />
                        </Paper>
                    </Container >

                )}
        </LoginContext.Consumer>
    )
}