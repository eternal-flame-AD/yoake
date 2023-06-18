import { makeJSONRequest } from "./request";

/*
pub struct MedicationLog {
    pub uuid: String,
    pub med_uuid: String,
    pub dosage: i32,
    pub time_actual: chrono::NaiveDateTime,
    pub time_expected: chrono::NaiveDateTime,
    pub dose_offset: f32,
    pub created: chrono::NaiveDateTime,
    pub updated: chrono::NaiveDateTime,
}

*/

export interface MedicationLog {
    uuid: string,
    med_uuid: string,
    dosage: number,
    time_actual: string,
    time_expected: string,
    dose_offset: number,
    created: string,
    updated: string,
}

export async function projectNextDose(med_uuid: string): Promise<MedicationLog> {
    const url = `/api/med/directive/${med_uuid}/project_next_dose`;
    const method = "GET";
    return (await makeJSONRequest<MedicationLog>(url, method)).data;
}

export interface GetMedicationLogParams {
    until?: Date
    limit: number,
}

export async function getMedicationLog(med_uuid: string, params: GetMedicationLogParams): Promise<MedicationLog[]> {
    let url = `/api/med/directive/${med_uuid}/log?limit=${params.limit}`;
    if (params.until) {
        url += `&until=${params.until.toISOString()}`;
    }
    const method = "GET";
    return (await makeJSONRequest<MedicationLog[]>(url, method)).data;
}

export async function postMedicationLog(med: MedicationLog): Promise<MedicationLog> {
    const uri = `/api/med/directive/${med.med_uuid}/log`;
    const method = "POST";
    const body = med;
    return (await makeJSONRequest<MedicationLog>(uri, method, body)).data;
}


export async function deleteMedicationLog(med_uuid: string, uuid: string): Promise<void> {
    const url = `/api/med/directive/${med_uuid}/log/${uuid}`;
    const method = "DELETE";

    await makeJSONRequest<void>(url, method);
}