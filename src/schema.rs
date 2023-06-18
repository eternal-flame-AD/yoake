// @generated automatically by Diesel CLI.

diesel::table! {
    medication_logs (uuid) {
        uuid -> Text,
        med_uuid -> Text,
        dosage -> Integer,
        time_actual -> Timestamp,
        time_expected -> Timestamp,
        dose_offset -> Float,
        created -> Timestamp,
        updated -> Timestamp,
    }
}

diesel::table! {
    medications (uuid) {
        uuid -> Text,
        name -> Text,
        dosage -> Integer,
        dosage_unit -> Text,
        period_hours -> Integer,
        flags -> Text,
        options -> Text,
        created -> Timestamp,
        updated -> Timestamp,
    }
}

diesel::table! {
    sessions (uuid) {
        uuid -> Text,
        expiry -> Timestamp,
        content -> Text,
    }
}

diesel::joinable!(medication_logs -> medications (med_uuid));

diesel::allow_tables_to_appear_in_same_query!(
    medication_logs,
    medications,
    sessions,
);
