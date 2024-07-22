import sys
import argparse
import requests
import pycobertura

DEFAULT_RED_LIMIT=50
DEFAULT_GREEN_LIMIT=65
DEFAULT_LABEL = "Coverage"

def check_valid_percentage_range(percentage):
    return 0 <= percentage <= 100

def get_limit_values(red_limit, green_limit):
    """It returns the red and green limits that defines the percentage values in which the coverage change color from
       red to orange and from orange to green"""
    red_limit = red_limit or DEFAULT_RED_LIMIT
    green_limit = green_limit or DEFAULT_GREEN_LIMIT

    if not check_valid_percentage_range(red_limit) or not check_valid_percentage_range(green_limit):
        raise ValueError("Invalid value for Red Limit or Green Limit. They should be between 0 and 100")
    if red_limit >= green_limit:
        raise ValueError(f"Invalid limit value. Green limit ({green_limit}) should be greater than Red Limit ({red_limit}) ")
    return [red_limit, green_limit]

def parse_coverage_from_report(report_path):
    """It tries to open the coverage xml report and parses it to get the line rate. It returns the line coverage percentage"""
    try:
        coverage_report = pycobertura.Cobertura(report_path)
        return round(coverage_report.line_rate() * 100, 2)
    except Exception:
        raise ValueError("Invalid file path or invalid file format")

def parse_coverage(coverage):
    """Given a coverage percentage string, it parses it and checks if it is a valid percentage value.
       It returns the percentage as a float"""
    percentage = coverage
    if not coverage:
        raise ValueError("Percentage should not be empty")
    if coverage[-1] == "%":
        percentage = coverage[:-1]
    # we try to convert it to float
    try:
        percentage = float(percentage)
        if not check_valid_percentage_range(percentage):
            raise ValueError("Invalid percentage value")
        return percentage
    except ValueError:
        raise ValueError(f"Invalid percentage data provided: {coverage}. The value is not a valid percentage number")

def get_color(percentage, limits):
    """Returns the corresponding color depending on the percentage value and the corresponding limits"""
    red_limit, green_limit = limits
    if percentage >= green_limit:
        return "green"
    elif red_limit <= percentage < green_limit:
        return "orange"
    return "red"

def get_label(label):
    """Returns the label by replacing the spaces with _"""
    if not label:
        return DEFAULT_LABEL
    return label.replace(" ","_")

def get_destination_path(path):
    if not path:
        # We will save the image in the local directory
        return "coverage.svg"
    return path

def download_image(label, percentage, color, path):
    """It downloads the coverage badge generated by img.shields.io"""
    url = f"https://img.shields.io/badge/{label}-{percentage}%25-{color}"
    response = requests.get(url)
    if response.status_code >= 400:
        raise RuntimeError("An error ocurred while getting the badge")
    with open(path, mode="wb") as file:
        file.write(response.content)

def parse_opt():
    parser = argparse.ArgumentParser(prog='Generate Badge', description='This script generates a coverage badge')
    parser.add_argument('--coverage', type=str, help='Coverage percentage. If defined, input-report will be ignored. Example: 75%%')
    parser.add_argument('--input-report', type=str, default="coverage.xml", help="Path to the xml coverage report")
    parser.add_argument('--label', type=str, default=DEFAULT_LABEL, help='Left side label')
    parser.add_argument('--output', type=str, help="Path to save the badge created")
    parser.add_argument('--red-limit', type=float, default=DEFAULT_RED_LIMIT, help="Limit value to red status. It indicates the limit value in which the status change from red to orange")
    parser.add_argument('--green-limit', type=float, default=DEFAULT_GREEN_LIMIT, help="Limit value to green status. It indicates the limit value in which the status change from orange to green")

    opt = parser.parse_args()
    return vars(opt)

def run(arguments):
    coverage_input = arguments.get("coverage", "")
    coverage_report_path = arguments.get("input_report", "")
    if coverage_input:
        percentage = parse_coverage(coverage_input)
    else:
        percentage = parse_coverage_from_report(coverage_report_path)

    limits = get_limit_values(arguments.get("red_limit", 0), arguments.get("green_limit", 0))
    color = get_color(percentage, limits)
    label = get_label(arguments.get("label", ""))
    destination_path = get_destination_path(arguments.get("output", ""))
    download_image(label, percentage, color, destination_path)

if __name__ == '__main__':
    opt = parse_opt()
    run(opt)