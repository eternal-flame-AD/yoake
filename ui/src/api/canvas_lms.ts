import { makeJSONRequest } from "./request";


export interface Grading {
    name: string,
    course_name: string,
    course_code: string,

    assignment_id: string,
    assignment_legacy_id: string,
    assignment_url: string,
    submission_id: string,
    submission_legacy_id: string,
    course_id: string,
    course_legacy_id: string,

    due_at: string | null,
    state: string,
    score: number | null,
    entered_score: number | null,
    possible_points: number,
    grade: string,
    grade_hidden: boolean,
    entered_grade: string,

    graded_at: string | null,
    posted_at: string | null,
}

export interface GetGradingsResponse {
    last_updated: string,
    response: Grading[],
}

export async function getGradings(force: boolean): Promise<GetGradingsResponse> {
    let ret = await makeJSONRequest<GetGradingsResponse>("/api/canvas_lms/grades" + (force ? "?force_refresh=true" : ""), "GET");
    if (ret.status != "Ok") {
        throw new Error(ret.message);
    }
    return ret.data;
}