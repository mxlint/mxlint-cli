package communitycommons;

import com.mendix.core.Core;
import com.mendix.core.CoreException;
import com.mendix.core.conf.RuntimeVersion;
import com.mendix.core.objectmanagement.member.MendixBoolean;
import com.mendix.integration.WebserviceException;
import com.mendix.systemwideinterfaces.core.IContext;
import com.mendix.systemwideinterfaces.core.IMendixObject;
import com.mendix.systemwideinterfaces.core.ISession;
import com.mendix.systemwideinterfaces.core.IUser;
import communitycommons.proxies.LogNodes;
import static communitycommons.proxies.constants.Constants.getMergeMultiplePdfs_MaxAtOnce;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.URL;
import java.net.URLConnection;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Iterator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.ThreadFactory;
import java.util.concurrent.atomic.AtomicLong;
import java.util.stream.Collectors;
import org.apache.commons.io.IOUtils;
import org.apache.pdfbox.multipdf.Overlay;
import org.apache.pdfbox.multipdf.PDFMergerUtility;
import org.apache.pdfbox.pdmodel.PDDocument;
import system.proxies.FileDocument;
import system.proxies.Language;

public class Misc {

	private static final String LOGNODE = LogNodes.CommunityCommons.name();
	private static boolean UNDER_TEST = false;

	static {
		try {
			// if this succeeds, we have a runtime
			Logging.createLogNode(LOGNODE);
		} catch (NullPointerException npe) {
			UNDER_TEST = true;
		}
	}

	public abstract static class IterateCallback<T1, T2> {

		boolean start = false;
		boolean stop = false;
		private Iterator<T1> mapIter;

		public abstract void hit(T1 key, T2 value) throws Exception;

		public void exit() {
			stop = true;
		}

		public void remove() {
			mapIter.remove();
		}

		synchronized void runOn(Map<T1, T2> map) throws Exception {
			if (start) {
				throw new IllegalMonitorStateException();
			}
			start = true;

			try {
				this.mapIter = map.keySet().iterator();

				while (mapIter.hasNext()) {
					T1 key = mapIter.next();
					T2 value = map.get(key);

					hit(key, value);

					if (stop) {
						break;
					}
				}
			} finally {
				//reset state to allow reuse, even when exceptions occur
				mapIter = null;
				stop = false;
				start = false;
			}
		}
	}

	/**
	 *
	 * @param <E> Any Java enumeration, such as a proxy class for a Mendix enumeration
	 * @param enumClass For instance: LogLevel.class
	 * @param value The value to look up in the enumeration
	 * @return An Optional of the requested enumeration. Will have the requested value if found,
	 * Optional.empty() otherwise.
	 */
	public static <E extends Enum<E>> Optional<E> enumFromString(final Class<E> enumClass, final String value) {
		try {
			if (value != null) {
				return Optional.of(E.valueOf(enumClass, value));
			}
		} catch (IllegalArgumentException iae) {
			if (!UNDER_TEST) {
				Logging.warn(LOGNODE, String.format("No enumeration with value %s found", value));
			}
		}
		return Optional.empty();
	}

	/**
	 * Because you cannot remove items from a map while iteration, this function is introduced. In
	 * the callback, you can use this.remove() or this.exit() to either remove or break the loop.
	 * Use return; to continue
	 *
	 * @throws Exception
	 */
	public static <A, B> void iterateMap(Map<A, B> map, IterateCallback<A, B> callback) throws Exception {
		//http://marxsoftware.blogspot.com/2008/04/removing-entry-from-java-map-during.html
		if (map == null || callback == null) {
			throw new IllegalArgumentException();
		}

		callback.runOn(map);
	}

	public static String getApplicationURL() {
		final String applicationURL = UNDER_TEST ? "http://localhost:8080/" : Core.getConfiguration().getApplicationRootUrl();
		return StringUtils.removeEnd(applicationURL, "/");
	}

	public static String getRuntimeVersion() {
		RuntimeVersion runtimeVersion = RuntimeVersion.getInstance();
		return runtimeVersion.toString();
	}

	public static String getModelVersion() {
		return Core.getModelVersion();
	}

	public static void throwException(String message) throws UserThrownException {
		throw new UserThrownException(message);
	}

	public static void throwWebserviceException(String faultstring) throws WebserviceException {
		throw new WebserviceException(WebserviceException.clientFaultCode, faultstring);
	}

	public static String retrieveURL(String url, String postdata) throws Exception {
		// Send data, appname
		URLConnection conn = new URL(url).openConnection();

		conn.setDoInput(true);
		conn.setDoOutput(true);
		conn.setRequestProperty("Content-Type", "application/x-www-form-urlencoded");

		if (postdata != null) {
			try (
				OutputStream os = conn.getOutputStream()) {
				IOUtils.copy(new ByteArrayInputStream(postdata.getBytes(StandardCharsets.UTF_8)), os);
			}
		}

		String result;
		try (
			InputStream is = conn.getInputStream()) {
			// Get the response
			result = IOUtils.toString(is, StandardCharsets.UTF_8);
		}

		return result;
	}

	public static Boolean duplicateFileDocument(IContext context, IMendixObject toClone, IMendixObject target) throws Exception {
		if (toClone == null || target == null) {
			throw new Exception("No file to clone or to clone into provided");
		}

		MendixBoolean hasContents = (MendixBoolean) toClone.getMember(context, FileDocument.MemberNames.HasContents.toString());
		if (!hasContents.getValue(context)) {
			return false;
		}

		try (
			InputStream inputStream = Core.getFileDocumentContent(context, toClone)) {
			Core.storeFileDocumentContent(context, target, (String) toClone.getValue(context, system.proxies.FileDocument.MemberNames.Name.toString()), inputStream);
		}

		return true;
	}

	public static Boolean duplicateImage(IContext context, IMendixObject toClone, IMendixObject target, int thumbWidth, int thumbHeight) throws Exception {
		if (toClone == null || target == null) {
			throw new Exception("No file to clone or to clone into provided");
		}

		MendixBoolean hasContents = (MendixBoolean) toClone.getMember(context, FileDocument.MemberNames.HasContents.toString());
		if (!hasContents.getValue(context)) {
			return false;
		}

		try (
			InputStream inputStream = Core.getImage(context, toClone, false)) {
			Core.storeImageDocumentContent(context, target, inputStream, thumbWidth, thumbHeight);

		}

		return true;
	}

	public static Boolean storeURLToFileDocument(IContext context, String url, IMendixObject __document, String filename) throws IOException {
		if (__document == null || url == null || filename == null) {
			throw new IllegalArgumentException("No document, filename or URL provided");
		}

		final int MAX_REMOTE_FILESIZE = 1024 * 1024 * 200; //maximum of 200 MB
		try {
			URL imageUrl = new URL(url);
			URLConnection connection = imageUrl.openConnection();
			//we connect in 20 seconds or not at all
			connection.setConnectTimeout(20000);
			connection.setReadTimeout(20000);
			connection.connect();

			int contentLength = connection.getContentLength();

			//check on forehand the size of the remote file, we don't want to kill the server by providing a 3 terabyte image.
			Logging.trace(LOGNODE, String.format("Remote filesize: %d", contentLength));

			if (contentLength > MAX_REMOTE_FILESIZE) { //maximum of 200 mb
				throw new IllegalArgumentException(String.format("Wrong filesize of remote url: %d (max: %d)", contentLength, MAX_REMOTE_FILESIZE));
			}

			InputStream fileContentIS;
			try (InputStream connectionInputStream = connection.getInputStream()) {
				if (contentLength >= 0) {
					fileContentIS = connectionInputStream;
				} else { // contentLength is negative or unknown
					Logging.trace(LOGNODE, String.format("Unknown content length; limiting to %d", MAX_REMOTE_FILESIZE));
					byte[] outBytes = new byte[MAX_REMOTE_FILESIZE];
					int actualLength = IOUtils.read(connectionInputStream, outBytes, 0, MAX_REMOTE_FILESIZE);
					fileContentIS = new ByteArrayInputStream(Arrays.copyOf(outBytes, actualLength));
				}
				Core.storeFileDocumentContent(context, __document, filename, fileContentIS);
			}
		} catch (IOException ioe) {
			Logging.error(LOGNODE, String.format("A problem occurred while reading from URL %s: %s", url, ioe.getMessage()));
			throw ioe;
		}

		return true;
	}

	public static Long getFileSize(IContext context, IMendixObject document) {
		final int BUFFER_SIZE = 4096;
		long size = 0;

		if (context != null) {
			byte[] buffer = new byte[BUFFER_SIZE];

			try (
				InputStream inputStream = Core.getFileDocumentContent(context, document)) {
				int i;
				while ((i = inputStream.read(buffer)) != -1) {
					size += i;
				}
			} catch (IOException e) {
				Logging.error(LOGNODE, "Couldn't determine filesize of FileDocument '" + document.getId());
			}
		}

		return size;
	}

	public static void delay(long delaytime) throws InterruptedException {
		Thread.sleep(delaytime);
	}

	public static IContext getContextFor(IContext context, String username, boolean sudoContext) {
		if (username == null || username.isEmpty()) {
			throw new RuntimeException("Assertion: No username provided");
		}

		if (username.equals(context.getSession().getUserName())) {
			return context;
		} else { // when it is a scheduled event, then get the right session.
			ISession session = getSessionFor(context, username);

			IContext c = session.createContext();
			if (sudoContext) {
				return c.createSudoClone();
			}

			return c;
		}
	}

	private static ISession getSessionFor(IContext context, String username) {
		ISession session = Core.getActiveSession(username);

		if (session == null) {
			IContext newContext = context.getSession().createContext().createSudoClone();
			newContext.startTransaction();
			try {
				session = initializeSessionForUser(newContext, username);
			} catch (CoreException e) {
				newContext.rollbackTransaction();

				throw new RuntimeException("Failed to initialize session for user: " + username + ": " + e.getMessage(), e);
			} finally {
				newContext.endTransaction();
			}
		}

		return session;
	}

	private static ISession initializeSessionForUser(IContext context, String username) throws CoreException {
		IUser user = Core.getUser(context, username);

		if (user == null) {
			throw new RuntimeException("Assertion: user with username '" + username + "' does not exist");
		}

		return Core.initializeSession(user, null);
	}

	public static Object executeMicroflowAsUser(IContext context,
		String microflowName, String username, Boolean sudoContext, Object... args) throws Exception {

		if (context == null) {
			throw new Exception("Assertion: No context provided");
		}
		if (microflowName == null || microflowName.isEmpty()) {
			throw new Exception("Assertion: No context provided");
		}
		if (!Core.getMicroflowNames().contains(microflowName)) {
			throw new Exception("Assertion: microflow not found: " + microflowName);
		}
		if (args.length % 2 != 0) {
			throw new Exception("Assertion: odd number of dynamic arguments provided, please name every argument: " + args.length);
		}

		Map<String, Object> params = new LinkedHashMap<String, Object>();
		for (int i = 0; i < args.length; i += 2) {
			if (args[i] != null) {
				params.put(args[i].toString(), args[i + 1]);
			}
		}

		IContext c = getContextFor(context, username, sudoContext);

		return Core.microflowCall(microflowName).withParams(params).execute(c);
	}

	//MWE: based on: http://download.oracle.com/javase/6/docs/api/java/util/concurrent/Executor.html
	static class MFSerialExecutor {

		private static MFSerialExecutor _instance = new MFSerialExecutor();

		private final AtomicLong tasknr = new AtomicLong();
		private final ExecutorService executor;

		public static MFSerialExecutor instance() {
			return _instance;
		}

		private MFSerialExecutor() {
			executor = Executors.newSingleThreadExecutor(new ThreadFactory() {

				//Default thread factory takes care of setting the proper thread context
				private final ThreadFactory defaultFactory = Executors.defaultThreadFactory();

				@Override
				public Thread newThread(Runnable runnable) {
					Thread t = defaultFactory.newThread(runnable);
					t.setPriority(Thread.MIN_PRIORITY);
					t.setName("CommunityCommons background pool executor thread");
					return t;
				}

			});
		}

		public void execute(final Runnable command) {
			if (command == null) {
				throw new NullPointerException("command");
			}

			final long currenttasknr = tasknr.incrementAndGet();
			Logging.debug(LOGNODE, "[RunMicroflowAsyncInQueue] Scheduling task #" + currenttasknr);

			executor.submit(new Runnable() {
				@Override
				public void run() {
					Logging.debug(LOGNODE, "[RunMicroflowAsyncInQueue] Running task #" + currenttasknr);
					try {
						command.run();
					} catch (RuntimeException e) {
						Logging.error(LOGNODE, "[RunMicroflowAsyncInQueue] Execution of task #" + currenttasknr + " failed: " + e.getMessage(), e);
						throw e; //single thread executor will continue, even if an exception is thrown.
					}
					Logging.debug(LOGNODE, "[RunMicroflowAsyncInQueue] Completed task #" + currenttasknr + ". Tasks left: " + (tasknr.get() - currenttasknr));
				}
			});
		}
	}

	public static Boolean runMicroflowAsyncInQueue(final String microflowName) {
		MFSerialExecutor.instance().execute(new Runnable() {
			@Override
			public void run() {
				try {
					Future<Object> future = Core.executeAsync(Core.createSystemContext(), microflowName, true, new HashMap<>()); //MWE: somehow, it only works with system context... well thats OK for now.
					future.get();
				} catch (CoreException | InterruptedException | ExecutionException e) {
					throw new RuntimeException("Failed to run Async: " + microflowName + ": " + e.getMessage(), e);
				}
			}
		});
		return true;
	}

	public static Boolean runMicroflowInBackground(final IContext context, final String microflowName,
		final IMendixObject paramObject) {

		MFSerialExecutor.instance().execute(new Runnable() {

			@Override
			public void run() {
				try {
					IContext c = Core.createSystemContext();
					if (paramObject != null) {
						Core.executeAsync(c, microflowName, true, paramObject).get(); //MWE: somehow, it only works with system context... well thats OK for now.
					} else {
						Core.executeAsync(c, microflowName, true, new HashMap<>()).get(); //MWE: somehow, it only works with system context... well thats OK for now.
					}
				} catch (CoreException | InterruptedException | ExecutionException e) {
					throw new RuntimeException("Failed to run Async: " + microflowName + ": " + e.getMessage(), e);
				}

			}

		});
		return true;
	}

	private interface IBatchItemHandler {

		void exec(IContext context, IMendixObject obj) throws Exception;

	}

	private static class BatchState {

		private int state = 0; //-1 = error, 1 = done.
		private final IBatchItemHandler callback;

		public BatchState(IBatchItemHandler callback) {
			this.callback = callback;
		}

		public void setState(int state) {
			this.state = state;
		}

		public int getState() {
			return state;
		}

		public void handle(IContext context, IMendixObject obj) throws Exception {
			callback.exec(context, obj);
		}
	}

	public static Boolean executeMicroflowInBatches(String xpath, final String microflow, int batchsize, boolean waitUntilFinished, boolean asc) throws CoreException, InterruptedException {
		Logging.debug(LOGNODE, "[ExecuteInBatches] Starting microflow batch '" + microflow + "...");

		return executeInBatches(xpath, new BatchState(new IBatchItemHandler() {

			@Override
			public void exec(IContext context, IMendixObject obj) throws Exception {
				Core.executeAsync(context, microflow, true, obj).get();
			}

		}), batchsize, waitUntilFinished, asc);
	}

	public static Boolean recommitInBatches(String xpath, int batchsize,
		boolean waitUntilFinished, Boolean asc) throws CoreException, InterruptedException {
		Logging.debug(LOGNODE, "[ExecuteInBatches] Starting recommit batch...");

		return executeInBatches(xpath, new BatchState(new IBatchItemHandler() {

			@Override
			public void exec(IContext context, IMendixObject obj) throws Exception {
				Core.commit(context, obj);
			}

		}), batchsize, waitUntilFinished, asc);
	}

	public static Boolean executeInBatches(String xpathRaw, BatchState batchState, int batchsize, boolean waitUntilFinished, boolean asc) throws CoreException, InterruptedException {
		String xpath = xpathRaw.startsWith("//") ? xpathRaw : "//" + xpathRaw;

		long count = Core.retrieveXPathQueryAggregate(Core.createSystemContext(), "count(" + xpath + ")");
		int loop = (int) Math.ceil(((float) count) / ((float) batchsize));

		Logging.debug(LOGNODE,
			"[ExecuteInBatches] Starting batch on ~ " + count + " objects divided over ~ " + loop + " batches. "
			+ (waitUntilFinished ? "Waiting until the batch has finished..." : "")
		);

		executeInBatchesHelper(xpath, batchsize, 0, batchState, count, asc);

		if (waitUntilFinished) {
			while (batchState.getState() == 0) {
				Thread.sleep(5000);
			}
			if (batchState.getState() == 1) {
				Logging.debug(LOGNODE, "[ExecuteInBatches] Successfully finished batch");
				return true;
			}
			Logging.error(LOGNODE, "[ExecuteInBatches] Failed to finish batch. Please check the application log for more details.");
			return false;
		}

		return true;
	}

	static void executeInBatchesHelper(final String xpath, final int batchsize, final long last, final BatchState batchState, final long count, final boolean asc) {
		MFSerialExecutor.instance().execute(new Runnable() {

			@Override
			public void run() {
				try {
					Thread.sleep(200);
					IContext c = Core.createSystemContext();

					List<IMendixObject> objects = Core.retrieveXPathQuery(c, xpath + (last > 0 ? "[id " + (asc ? "> " : "< ") + last + "]" : ""),
						batchsize,
						0,
						new HashMap<String, String>() {
						{
							put("id", asc ? "asc" : "desc");
						}
					}
					);

					//no new objects found :)
					if (objects.isEmpty()) {
						Logging.debug(LOGNODE, "[ExecuteInBatches] Succesfully finished batch on ~" + count + " objects.");
						batchState.setState(1);
					} else {

						//process objects
						for (IMendixObject obj : objects) {
							batchState.handle(c, obj);
						}

						//invoke next batch
						executeInBatchesHelper(xpath, batchsize, objects.get(objects.size() - 1).getId().toLong(), batchState, count, asc);
					}
				} catch (Exception e) {
					batchState.setState(-1);
					throw new RuntimeException("[ExecuteInBatches] Failed to run in batch: " + e.getMessage(), e);
				}
			}

		});
	}

	/**
	 * Tests if two objects are equal with throwing unecessary null pointer exceptions.
	 *
	 * This is almost the most stupid function ever, since it should be part of Java itself.
	 *
	 * @param left
	 * @param right
	 * @return
	 * @deprecated Native Java function Objects.equals() is available since Java 7
	 */
	@Deprecated
	public static boolean objectsAreEqual(Object left, Object right) {
		if (left == null && right == null) {
			return true;
		}
		if (left == null || right == null) {
			return false;
		}
		return left.equals(right);
	}

	/**
	 * Get the default language
	 *
	 * @param context
	 * @return The default language
	 * @throws CoreException
	 */
	public static Language getDefaultLanguage(IContext context) throws CoreException {
		String languageCode = Core.getDefaultLanguage().getCode();
		List<Language> languageList = Language.load(context, "[Code = '" + languageCode + "']");
		if (languageList == null || languageList.isEmpty()) {
			throw new RuntimeException("No language found for default language constant value " + languageCode);
		}
		return languageList.get(0);
	}

	public static boolean mergePDF(IContext context, List<FileDocument> documents, IMendixObject mergedDocument) throws IOException {
		if (getMergeMultiplePdfs_MaxAtOnce() <= 0 || documents.size() <= getMergeMultiplePdfs_MaxAtOnce()) {

			List<InputStream> sources = new ArrayList<>();
			try (ByteArrayOutputStream out = new ByteArrayOutputStream()) {
				PDFMergerUtility mergePdf = new PDFMergerUtility();

				for (FileDocument file : documents) {
					InputStream content = Core.getFileDocumentContent(context, file.getMendixObject());
					sources.add(content);
				}
				mergePdf.addSources(sources);
				mergePdf.setDestinationStream(out);
				mergePdf.mergeDocuments(null);

				Core.storeFileDocumentContent(context, mergedDocument, new ByteArrayInputStream(out.toByteArray()));

				out.reset();
				documents.clear();
			} catch (IOException e) {
				throw new RuntimeException("Failed to merge documents" + e.getMessage(), e);
			} finally { // We cannot use try-with-resources because streams would be prematurely closed
				for (InputStream is : sources) {
					is.close();
				}
			}

			return true;
		} else {
			throw new IllegalArgumentException("MergeMultiplePDFs: you cannot merge more than " + getMergeMultiplePdfs_MaxAtOnce()
				+ " PDF files at once. You are trying to merge " + documents.size() + " PDF files.");
		}
	}

	/**
	 * Overlay a generated PDF document with another PDF (containing the company stationary for
	 * example)
	 *
	 * @param context
	 * @param generatedDocumentMendixObject The document to overlay
	 * @param overlayMendixObject The document containing the overlay
	 * @param onTopOfContent if true, puts overlay position in the foreground, otherwise in the
	 * background
	 * @return boolean
	 * @throws IOException
	 */
	public static boolean overlayPdf(IContext context, IMendixObject generatedDocumentMendixObject, IMendixObject overlayMendixObject, boolean onTopOfContent) throws IOException {
		Logging.trace(LOGNODE, "Retrieve generated document");
		try (
			PDDocument inputDoc = PDDocument.load(Core.getFileDocumentContent(context, generatedDocumentMendixObject));
			PDDocument overlayDoc = PDDocument.load(Core.getFileDocumentContent(context, overlayMendixObject));
			ByteArrayOutputStream baos = new ByteArrayOutputStream()) {
			Logging.trace(LOGNODE, "Overlay PDF start, retrieve overlay PDF");

			Logging.trace(LOGNODE, "Perform overlay");
			Overlay overlay = new Overlay();
			overlay.setInputPDF(inputDoc);
			overlay.setDefaultOverlayPDF(overlayDoc);
			if (onTopOfContent == true) {
				overlay.setOverlayPosition(Overlay.Position.FOREGROUND);
			} else {
				overlay.setOverlayPosition(Overlay.Position.BACKGROUND);
			}

			Logging.trace(LOGNODE, "Save result in output stream");

			overlay.overlay(new HashMap<>()).save(baos);

			Logging.trace(LOGNODE, "Duplicate result in input stream");
			try (InputStream overlayedContent = new ByteArrayInputStream(baos.toByteArray())) {
				Logging.trace(LOGNODE, "Store result in original document");
				Core.storeFileDocumentContent(context, generatedDocumentMendixObject, overlayedContent);
			}
		}

		Logging.trace(LOGNODE, "Overlay PDF end");
		return true;
	}

	/**
	 * Get the Cloud Foundry Instance Index (0 for leader and >0 for slave)
	 *
	 * @return CF_INSTANCE_INDEX environment variable if available, otherwise -1
	 */
	public static long getCFInstanceIndex() {
		long cfInstanceIndex = -1L;

		try {
			cfInstanceIndex = Long.parseLong(System.getenv("CF_INSTANCE_INDEX"));
		} catch (SecurityException securityException) {
			Logging.info(LOGNODE, "GetCFInstanceIndex: Could not access environment variable CF_INSTANCE_INDEX, permission denied. Value of -1 is returned.");
		} catch (NumberFormatException numberFormatException) {
			Logging.info(LOGNODE, "GetCFInstanceIndex: Could not parse value of environment variable CF_INSTANCE_INDEX as Long. Value of -1 is returned.");
		} catch (NullPointerException nullPointerException) {
			Logging.info(LOGNODE, "GetCFInstanceIndex: Could not find value for environment variable CF_INSTANCE_INDEX. This could indicate a local deployment is running. Value of -1 is returned.");
		}

		return cfInstanceIndex;
	}

	/**
	 * Returns the top n items of a List
	 *
	 * @param <T> the type of the list elements
	 * @param objects the list to take the top n items from
	 * @param top the number of items to take
	 * @return the sublist of <pre>objects</pre> with max
	 * <pre>top</pre> elements
	 */
	public static <T> List<T> listTop(List<T> objects, int top) {
		return objects.stream()
			.limit(top)
			.collect(Collectors.toList());
	}
}
