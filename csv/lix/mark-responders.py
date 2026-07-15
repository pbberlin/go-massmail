from pathlib import Path
import csv


invitation = Path("invitation32_cleaned.csv")
responded  = Path("responses.csv")

out = Path("reminder.csv")


userIdsToSkip = set()

with responded.open("r", encoding="utf-8", newline="") as fileHandle:
    reader = csv.DictReader(fileHandle, delimiter=";")
    for idx1, row in enumerate(reader):
        if idx1%50 == 0:
            print(f" {idx1:5} responded:   {row["user_id"]}  {row["closing_time"]}")
        userIdsToSkip.add(row["user_id"])




print()

cntr = 0
with invitation.open("r", encoding="utf-8", newline="") as inputHandle:

    reader = csv.DictReader(inputHandle, delimiter=";")

    fieldNames = reader.fieldnames

    with out.open("w", encoding="utf-8", newline="") as outputHandle:
        writer = csv.DictWriter(
            outputHandle,
            fieldnames=fieldNames,
            delimiter=";"
        )

        writer.writeheader()

        for idx1, row in enumerate(reader):
            skipValue = (
                row["userid"] in userIdsToSkip
            )
            if skipValue:
                row["skip"] = "responded"
                cntr += 1
                if idx1%50 == 0 or True:
                    print(f" {cntr:5}   marking:   {row["userid"]}  {row["vorname"]}")

            writer.writerow(row)


