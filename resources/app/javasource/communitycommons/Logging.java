package communitycommons;

import com.mendix.core.Core;
import com.mendix.logging.ILogNode;
import communitycommons.proxies.LogLevel;
import communitycommons.proxies.LogNodes;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;

public class Logging {

	private static Map<String, Long> timers = new HashMap<String, Long>();

	public static void trace(String lognode, String message) {
		log(lognode, LogLevel.Trace, message, null);
	}

	public static void info(String lognode, String message) {
		log(lognode, LogLevel.Info, message, null);
	}

	public static void debug(String lognode, String message) {
		log(lognode, LogLevel.Debug, message, null);
	}

	public static void warn(String lognode, String message, Throwable e) {
		log(lognode, LogLevel.Warning, message, e);
	}

	public static void warn(String lognode, String message) {
		warn(lognode, message, null);
	}

	public static void error(String lognode, String message, Throwable e) {
		log(lognode, LogLevel.Error, message, e);
	}

	public static void error(String lognode, String message) {
		error(lognode, message, null);
	}

	public static void critical(String lognode, String message, Throwable e) {
		log(lognode, LogLevel.Critical, message, e);
	}

	public static void log(String lognode, LogLevel loglevel, String message, Throwable e) {
		ILogNode logger = createLogNode(lognode);
		switch (loglevel) {
			case Critical:
				logger.critical(message, e);
				break;
			case Warning:
				logger.warn(message, e);
				break;
			case Debug:
				logger.debug(message);
				break;
			case Error:
				logger.error(message, e);
				break;
			case Info:
				logger.info(message);
				break;
			case Trace:
				logger.trace(message);
				break;
		}
	}

	public static Long measureEnd(String timerName, LogLevel loglevel,
		String message) {
		Long cur = new Date().getTime();
		if (!timers.containsKey(timerName)) {
			throw new IllegalArgumentException(String.format("Timer with key %s not found", timerName));
		}
		Long timeTaken = cur - timers.get(timerName);
		String time = String.format("%d", timeTaken);
		log(LogNodes.CommunityCommons.name(), loglevel, "Timer " + timerName + " finished in " + time + " ms. " + message, null);
		return timeTaken;
	}

	public static void measureStart(String timerName) {
		timers.put(timerName, new Date().getTime());
	}

	public static ILogNode createLogNode(String logNode) {
		return Core.getLogger(logNode);
	}
}
