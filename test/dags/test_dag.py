import sys
import os
import random

# Add project root to sys.path to find internal module
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../../')))

from internal.sdk.python import DAGBuilder, Task

# -----------------------------------------------------------
# User Logic (Pure Python)
# -----------------------------------------------------------

def extract_data():
    print("Extracting data from source...")
    print("Data extraction complete.")

def validate_data():
    print("Validating extracted data...")
    print("Validation passed.")

def check_data_quality() -> str:
    """Simulate quality check - returns branch name"""
    score = random.randint(0, 100)
    print(f"Data quality score: {score}")
    if score >= 70:
        return "high_quality"
    return "low_quality"

def process_high_quality():
    print("Processing high quality data - full pipeline")

def process_low_quality():
    print("Processing low quality data - applying corrections")

def clean_data():
    print("Cleaning and normalizing data...")

def load_data():
    print("Loading processed data to destination...")
    print("Pipeline complete!")

def main():
    (DAGBuilder("python-etl-pipeline")
        # Sequential tasks: extract -> validate
        .add_sequential(
            Task(name="extract", fn=extract_data),
            Task(name="validate", fn=validate_data),
        )
        # Conditional branching based on data quality
        .add_branch("check_quality", check_data_quality, {
            "high_quality": [
                Task(name="process_high", fn=process_high_quality),
            ],
            "low_quality": [
                Task(name="process_low", fn=process_low_quality),
                Task(name="clean", fn=clean_data),
            ],
        })
        # Join: wait for either branch to complete
        .add_join("load", load_data, ["process_high", "clean"])
        .serve()
    )

if __name__ == "__main__":
    main()
