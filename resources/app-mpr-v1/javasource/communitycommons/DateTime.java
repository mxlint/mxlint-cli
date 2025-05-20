package communitycommons;

import communitycommons.proxies.DatePartSelector;
import static communitycommons.proxies.DatePartSelector.day;
import static communitycommons.proxies.DatePartSelector.month;
import static communitycommons.proxies.DatePartSelector.year;
import java.util.Date;

import java.time.LocalDate;
import java.time.Period;
import java.time.ZoneId;
import java.util.Calendar;

public class DateTime {

	/**
	 * @author mwe
	 * @author res
	 * @param firstDate The begin of the period
	 * @param compareDate The end of the period
	 * @return The period between the firstDate in the system default timezone, and the compareDate in the system
	 * default timezone as a Java Period
	 *
	 * Code is based on http://stackoverflow.com/questions/1116123/how-do-i-calculate-someones-age-in-java
	 *
	 * Adjusted to Java 8 APIs (April, 2019)
	 */
	public static Period periodBetween(Date firstDate, Date compareDate) {
		return Period.between(toLocalDate(firstDate), toLocalDate(compareDate));
	}

	private static LocalDate toLocalDate(Date someDate) {
		return someDate.toInstant()
				.atZone(ZoneId.systemDefault())
				.toLocalDate();
	}

	public static long dateTimeToInteger(Date date, DatePartSelector selectorObj) {
		Calendar newDate = Calendar.getInstance();
		newDate.setTime(date);
		int value = -1;
		switch (selectorObj) {
			case year:
				value = newDate.get(Calendar.YEAR);
				break;
			case month:
				value = newDate.get(Calendar.MONTH) + 1;
				break; // Return starts at 0
			case day:
				value = newDate.get(Calendar.DAY_OF_MONTH);
				break;
			default:
				break;
		}
		return value;
	}

	public static long dateTimeToLong(Date date) {
		return date.getTime();
	}

	public static Date longToDateTime(Long value) {
		return new Date(value);
	}
}
