import datetime as dt
import importlib.util
import pathlib
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location(
    "leamout_live_runner", pathlib.Path(__file__).with_name("run.py")
)
live = importlib.util.module_from_spec(spec)
spec.loader.exec_module(live)


class LiveRunnerTests(unittest.TestCase):
    def test_api_origin(self):
        for origin in ("https://example.com", "http://localhost:8080"):
            self.assertEqual(live.validate_origin(origin), origin)
        for origin in ("http://example.com", "https://user:pass@example.com",
                       "https://example.com/path", "ftp://example.com"):
            with self.subTest(origin=origin), self.assertRaises(live.AcceptanceError):
                live.validate_origin(origin)

    def test_media_intervals(self):
        now = dt.datetime.now(dt.timezone.utc)
        def at(seconds):
            return (now + dt.timedelta(seconds=seconds)).isoformat()
        base = {"media_started_at": at(0), "media_ended_at": at(2), "ended_at": at(3)}
        live.validate_media(base)
        for change in ({"media_started_at": None}, {"media_ended_at": at(0)},
                       {"media_ended_at": at(4)}, {"ended_at": None}):
            with self.subTest(change=change), self.assertRaises(live.AcceptanceError):
                live.validate_media({**base, **change})

    def test_nil_uuid_rejected(self):
        with patch.dict("os.environ", {"TEST_LIVE_UUID": "00000000-0000-0000-0000-000000000000"}):
            with self.assertRaises(live.AcceptanceError):
                live.require_uuid("TEST_LIVE_UUID")

    def test_opt_in_required(self):
        args = live.argparse.Namespace(direction="inbound", timeout=120, require_tool=False,
                                       require_barge_in=False, require_recording=False)
        with patch.dict("os.environ", {"LEAMOUT_LIVE_CONFIRM": ""}):
            with self.assertRaises(live.AcceptanceError):
                live.Runner(args)


if __name__ == "__main__":
    unittest.main()
