# Badge Generator
Script that generates a coverage badge. You can customize the following parameters:

- The left label that will appear in the badge
- The color range limits. By default, the badge will be red between 0-50%, orange between 50-65% and green if higher than 65
- The output file name

The coverage percentage be passed in two different ways:

- By giving the desired value using the flag `--coverage`
- By indicating the path to the XML coverage report using the flag `--input-report`

Installation
------------
```bash
# Generate python virtual environment and activate it
python3 -m venv venv
source venv/bin/activate

# Install requirements
pip install -r requirements.txt
```

Examples
------------

```bash
# Run help
python generate_badge.py --help

# Generate badge with default values and indicating the coverage percentage
python generate_badge.py --coverage 75

# Generate with xml report file
python generate_badge.py --input-report coverage.xml

# Generate badge with multiple parameters
python generate_badge.py --label "API coverage" --input-report coverage.xml --output coverage_api.svg --red-limit 55 --green-limit 70
```
