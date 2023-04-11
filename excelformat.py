import pandas as pd

# Load the original table into a DataFrame
df = pd.read_csv('athletes.csv')

# Split the table into multiple tables based on a space character
tables = []
start_index = 1
for i, row in df.iterrows():
    if row.isnull().all():
        # Found a space, so split the table at this point
        tables.append(df.iloc[start_index:i])
        start_index = i + 1

# Append the last table
tables.append(df.iloc[start_index:])

# Remove rows that start with a letter from each table
for table in tables:
    table.drop(table[table.iloc[:,0].str.match('[a-zA-Z]')].index, inplace=True)

# Write each table to a separate section in the same CSV file
with open('output_table.csv', 'w') as f:
    for i, table in enumerate(tables):
        # Write a section header
        f.write(f'Table {i+1}\n')
        # Write the table to the file (excluding the index column)
        table.to_csv(f, index=False)
        # Add a blank line between tables
        f.write('\n')