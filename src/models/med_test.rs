use super::med::*;

#[test]
pub fn test_parse_med() {
    let uuid_stub = "".to_string();
    let time_stub = chrono::Utc::now().naive_utc();
    let cases = vec![
        (
            "Atorvastatin 10mg QD",
            Medication {
                uuid: uuid_stub.clone(),
                name: "Atorvastatin".to_string(),
                dosage: 10,
                dosage_unit: "mg".to_string(),
                period_hours: 24,
                flags: "".to_string(),
                options: "".to_string(),
                created: time_stub.clone(),
                updated: time_stub.clone(),
            },
        ),
        (
            "Something 10mg tid adlib",
            Medication {
                uuid: uuid_stub.clone(),
                name: "Something".to_string(),
                dosage: 10,
                dosage_unit: "mg".to_string(),
                period_hours: 8,
                flags: "adlib".to_string(),
                options: "".to_string(),
                created: time_stub.clone(),
                updated: time_stub.clone(),
            },
        ),
        (
            "Metformin 500mg qHS",
            Medication {
                uuid: uuid_stub.clone(),
                name: "Metformin".to_string(),
                dosage: 500,
                dosage_unit: "mg".to_string(),
                period_hours: 24,
                flags: "qhs".to_string(),
                options: "".to_string(),
                created: time_stub.clone(),
                updated: time_stub.clone(),
            },
        ),
        (
            "Hydroxyzine 50mg qid prn sched(whole)",
            Medication {
                uuid: uuid_stub.clone(),
                name: "Hydroxyzine".to_string(),
                dosage: 50,
                dosage_unit: "mg".to_string(),
                period_hours: 6,
                flags: "prn".to_string(),
                options: "sched(whole)".to_string(),
                created: time_stub.clone(),
                updated: time_stub.clone(),
            },
        ),
    ];
    for (input, expected) in cases {
        let actual = input.parse::<Medication>().unwrap();
        assert_eq!(actual.name, expected.name);
        assert_eq!(actual.dosage, expected.dosage);
        assert_eq!(actual.dosage_unit, expected.dosage_unit);
        assert_eq!(actual.period_hours, expected.period_hours);
        assert_eq!(actual.flags, expected.flags);
        assert_eq!(actual.options, expected.options);
    }
}
