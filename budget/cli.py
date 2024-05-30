import argparse
import logging
import os
from typing import Final

from budget.main import Args, main

logger = logging.getLogger(__name__)

SHEETS_RANGE_NAME: Final = "transactions"
MAPPING_RANGE_NAME: Final = "lookup"


def run() -> None:
    try:
        logger.info("Starting...")
        args = get_args()
        main(args)
        logger.info("Done")
    except KeyboardInterrupt:
        logger.info("Exiting...")
    except Args.Error as e:
        logger.error(e, exc_info=False)  # noqa: TRY400
    except Exception:
        logger.exception("An error occurred")


def get_args() -> Args:
    arg_parser = argparse.ArgumentParser(description="Budget CLI")
    _ = arg_parser.add_argument(
        "--simplefin-username",
        help="SimpleFin username",
        default=os.getenv("SIMPLE_FIN_USERNAME"),
    )
    _ = arg_parser.add_argument(
        "--simplefin-password",
        help="SimpleFin password",
        default=os.getenv("SIMPLE_FIN_PASSWORD"),
    )
    _ = arg_parser.add_argument(
        "--simplefin-access-url",
        help="SimpleFin access URL",
        default=os.getenv("SIMPLE_FIN_ACCESS_URL"),
    )
    _ = arg_parser.add_argument(
        "--paperless-url",
        help="Paperless URL",
        default=os.getenv("PAPERLESS_URL"),
    )
    _ = arg_parser.add_argument(
        "--paperless-token",
        help="Paperless token",
        default=os.getenv("PAPERLESS_TOKEN"),
    )
    _ = arg_parser.add_argument(
        "--google-credentials",
        help="Google credentials",
        default=os.getenv("GOOGLE_CREDENTIALS"),
    )
    _ = arg_parser.add_argument(
        "--sheets-spreadsheet-id",
        help="Google Sheets spreadsheet ID",
        default=os.getenv("SHEETS_SPREADSHEET_ID"),
    )
    _ = arg_parser.add_argument(
        "--sheets-range-name",
        help="Google Sheets range name",
        default=os.getenv("SHEETS_RANGE_NAME", SHEETS_RANGE_NAME),
    )
    _ = arg_parser.add_argument(
        "--mapping-range-name",
        help="Google Sheets mapping range name",
        default=os.getenv("MAPPING_RANGE_NAME", MAPPING_RANGE_NAME),
    )
    cli_args_dict: dict[str, str] = vars(arg_parser.parse_args())
    return Args(
        simplefin_username=cli_args_dict["simplefin_username"],
        simplefin_password=cli_args_dict["simplefin_password"],
        simplefin_access_url=cli_args_dict["simplefin_access_url"],
        paperless_url=cli_args_dict["paperless_url"],
        paperless_token=cli_args_dict["paperless_token"],
        google_credentials=cli_args_dict["google_credentials"],
        sheets_spreadsheet_id=cli_args_dict["sheets_spreadsheet_id"],
        sheets_range_name=cli_args_dict["sheets_range_name"],
        mapping_range_name=cli_args_dict["mapping_range_name"],
    )
