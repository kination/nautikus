import sys
import os
import random

# Add project root to sys.path to find internal module
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../../')))

from internal.sdk.python_sdk import serve

# -----------------------------------------------------------
# 1. User Logic (Pure Python)
# -----------------------------------------------------------

def check_system():
    print("Checking system resources...")
    print(f"System check passed. Random check code: {random.randint(1000, 9999)}")

def process_data():
    print("Processing data...")
    # Simulate some work
    data = [random.randint(1, 100) for _ in range(5)]
    print(f"Processed data: {data}")

def main():
    # Define DAG and task execution order naturally
    serve(
        dag_name="python-exclusive-dag",
        tasks=[check_system, process_data] # Sequential execution
    )

if __name__ == "__main__":
    main()