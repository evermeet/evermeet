import {
  format as _format,
  interval,
  isSameMonth,
  isSameYear,
  isSameDay,
  formatDuration,
  intervalToDuration,
} from "date-fns";
import {
  formatInTimeZone as _formatInTimeZone,
  getTimezoneOffset,
} from "date-fns-tz";

export { interval, isSameDay, isSameMonth, isSameYear } from "date-fns";

// https://github.com/date-fns/date-fns/blob/main/src/locale/cs/_lib/formatLong/index.ts
// https://github.com/date-fns/date-fns/blob/main/src/locale/en-US/_lib/formatLong/index.ts
// ....
const localeIntervals = [
  {
    locales: ["en", "en-US"],
    sameMonth: {
      full: "~EEEE~ - 'EEEE', ~MMMM do~ - 'do', ~y~",
    },
    sameYear: {
      full: "~EEEE~ - 'EEEE', ~MMMM do~ - 'MMMM do', ~y~",
    },
    default: {
      full: "~EEEE~ - 'EEEE', ~MMMM do, y~ - 'MMMM do, y'",
    },
  },
  {
    locales: ["es", "cs"],
    sameMonth: {
      full: "~EEEE~ - 'EEEE', ~d.~ - 'd.' ~MMMM yyyy~",
    },
    sameYear: {
      full: "~EEEE~ - 'EEEE', ~d. MMMM~ - 'd. MMMM' ~yyyy~",
    },
    default: {
      full: "~EEEE~ - 'EEEE', ~d. MMMM yyyy~ - 'd. MMMM yyyy'",
    },
  },
];

function findLocaleInterval(lang) {
  let conf = localeIntervals.find((li) => li.locales.includes(lang));
  if (!conf) {
    conf = localeIntervals.find((li) => li.locales.includes("en"));
  }
  return conf;
}

export function format(date, formatStr, opts = {}) {
  if (opts.timezone) {
    return _formatInTimeZone(date, opts.timezone, formatStr, opts);
  }
  return _format(date, formatStr, opts);
}

export function formatInTimeZone(d, tz, str, opts = {}) {
  return _formatInTimeZone(d, tz, str, opts);
}

export function formatInterval(i, type, { locale, localeInterval }) {
  let out = _format(i.start, localeInterval[type].full, { locale });
  out = _format(i.end, out.replace(/~/g, "'"), { locale });
  return out;
}

export function formatDateInterval(interval, opts = {}) {
  const localeInterval = findLocaleInterval(opts.locale?.code);
  const f = (d, str) => _formatInTimeZone(d, opts?.timezone, str, opts);
  const fi = (type) =>
    formatInterval(interval, type, Object.assign({}, opts, { localeInterval }));
  let out;
  if (isSameDay(interval.start, interval.end)) {
    // same day
    out = f(interval.start, `PPPP`);
  } else if (isSameMonth(interval.start, interval.end)) {
    // same month and year
    out = fi("sameMonth");
  } else if (isSameYear(interval.start, interval.end)) {
    // same year
    out = fi("sameYear");
  } else {
    // otherwise (cross-year events like 31. december - 1. january)
    out = `${f(interval.start, "iiii")} - ${f(interval.end, "iiii")}, ${f(interval.start, "d MMMM yyyy")} - ${f(interval.end, "d MMMM yyyy")}`;
  }
  return out;
}

export function formatTimeInterval(interval, opts = {}) {
  const user = opts.user;
  const f = (d, str) =>
    formatInTimeZone(d, opts.tz || opts.locale?.timezone, str, opts);
  let out;
  if (user?.preferences?.date?.hoursFormat === "12-hour") {
    out = `${f(interval.start, "h:mm aa")} - ${f(interval.end, "h:mm aa zzz")}`;
  } else if (user?.preferences?.date?.hoursFormat === "24-hour") {
    out = `${f(interval.start, "H:mm")} - ${f(interval.end, "H:mm zzz")}`;
  } else {
    out = `${f(interval.start, "p")} - ${f(interval.end, "p zzz")}`;
  }
  return out;
}

export function formatDurationInterval(interval, opts = {}) {
  return formatDuration(intervalToDuration(interval), opts);
}

export function timezonesOffset(time, a, b) {
  if (!b) {
    return 0;
  }
  const offset = getTimezoneOffset(a, time) - getTimezoneOffset(b, time);
  return offset;
}
