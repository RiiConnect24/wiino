# wiino

This code lets you extract variables from a Wii Number, verify if it's valid, or let you make your own with said variables.

The variables are:

- Hollywood ID (GPU chip in the Wii)
- Counter (increments every time you format your Wii, can range from 0-31)
- Hardware Model (RVL, RVT for dev hardware...)
- Area Code (Region - USA, EUR, JPN...)
- Checksum

The code currently works for these languages:

- C
- Golang
- Python
