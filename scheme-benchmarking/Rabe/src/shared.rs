use std::error::Error;
use std::fs::File;
use std::io::{BufReader, BufWriter};
use std::path::Path;

use csv::{ReaderBuilder, WriterBuilder};

pub fn test (){
    println!("testing");
}

pub fn update_csv(
    file_name: &str,
    index: &str,
    column: &str,
    value: &str,
) -> Result<(), Box<dyn Error>> {
    let file_path = Path::new(file_name);
    let mut records: Vec<Vec<String>> = Vec::new();
    let mut headers: Vec<String> = Vec::new();

    let file_exists = file_path.exists();

    // Read the CSV if it exists
    if file_exists {
        let file = File::open(file_path)?;
        let mut reader = ReaderBuilder::new()
            .has_headers(false)
            .from_reader(BufReader::new(file));

        for result in reader.records() {
            let record = result?;
            records.push(record.iter().map(|s| s.to_string()).collect());
        }

        if !records.is_empty() {
            headers = records[0].clone();
        }
    }

    // If the CSV is empty, initialize headers
    if records.is_empty() {
        headers.push("attributes".to_string());
        records.push(headers.clone());
    }

    // Check if the column exists
    let column_index = if let Some(pos) = headers.iter().position(|h| h == column) {
        pos
    } else {
        headers.push(column.to_string());
        let new_col_index = headers.len() - 1;
        records[0] = headers.clone();

        for row in records.iter_mut().skip(1) {
            while row.len() <= new_col_index {
                row.push("".to_string());
            }
        }

        new_col_index
    };

    // Find the row with the given index
    let mut row_index = None;
    for (i, record) in records.iter().enumerate().skip(1) {
        if !record.is_empty() && record[0] == index {
            row_index = Some(i);
            break;
        }
    }

    // If the row does not exist, add it
    let target_row_index = if let Some(idx) = row_index {
        idx
    } else {
        let mut new_row = Vec::new();
        new_row.push(index.to_string());
        new_row.resize(headers.len(), "".to_string());
        records.push(new_row);
        records.len() - 1
    };

    // Make sure the target row has enough columns
    while records[target_row_index].len() <= column_index {
        records[target_row_index].push("".to_string());
    }

    // Update the value
    records[target_row_index][column_index] = value.to_string();

    // Write the updated records back to the CSV
    let file = File::create(file_path)?;
    let mut writer = WriterBuilder::new()
        .has_headers(false)
        .from_writer(BufWriter::new(file));

    for record in records {
        writer.write_record(&record)?;
    }

    Ok(())
}
