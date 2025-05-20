package communitycommons;

import com.mendix.core.Core;
import com.mendix.core.CoreException;
import com.mendix.core.objectmanagement.member.MendixAutoNumber;
import com.mendix.core.objectmanagement.member.MendixDateTime;
import com.mendix.core.objectmanagement.member.MendixEnum;
import com.mendix.core.objectmanagement.member.MendixObjectReference;
import com.mendix.core.objectmanagement.member.MendixObjectReferenceSet;
import com.mendix.systemwideinterfaces.core.IContext;
import com.mendix.systemwideinterfaces.core.IMendixIdentifier;
import com.mendix.systemwideinterfaces.core.IMendixObject;
import com.mendix.systemwideinterfaces.core.IMendixObject.ObjectState;
import com.mendix.systemwideinterfaces.core.IMendixObjectMember;
import com.mendix.systemwideinterfaces.core.IMendixObjectMember.MemberState;
import com.mendix.systemwideinterfaces.core.meta.IMetaAssociation;
import com.mendix.systemwideinterfaces.core.meta.IMetaAssociation.AssociationType;
import com.mendix.systemwideinterfaces.core.meta.IMetaEnumValue;
import com.mendix.systemwideinterfaces.core.meta.IMetaEnumeration;
import com.mendix.systemwideinterfaces.core.meta.IMetaObject;
import com.mendix.systemwideinterfaces.core.meta.IMetaPrimitive;
import com.mendix.systemwideinterfaces.core.meta.IMetaPrimitive.PrimitiveType;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import system.proxies.FileDocument;

public class ORM {

	public static Long getGUID(IMendixObject item) {
		return item.getId().toLong();
	}

	public static String getOriginalValueAsString(IContext context, IMendixObject item,
		String member) {
		return String.valueOf(item.getMember(context, member).getOriginalValue(context));
	}

	public static boolean objectHasChanged(IMendixObject anyobject) {
		if (anyobject == null) {
			throw new IllegalArgumentException("The provided object is empty");
		}
		return anyobject.isChanged();
	}

	/**
	 * checks whether a certain member of an object has changed. If the objects itself is still new,
	 * we consider to be changes as well.
	 *
	 * @param item
	 * @param member
	 * @param context
	 * @return
	 */
	public static boolean memberHasChanged(IContext context, IMendixObject item, String member) {
		if (item == null) {
			throw new IllegalArgumentException("The provided object is empty");
		}
		if (!item.hasMember(member)) {
			throw new IllegalArgumentException("Unknown member: " + member);
		}
		return item.getMember(context, member).getState() == MemberState.CHANGED || item.getState() != ObjectState.NORMAL;
	}

	public static void deepClone(IContext c, IMendixObject source, IMendixObject target, String membersToSkip, String membersToKeep, String reverseAssociations, String excludeEntities, String excludeModules) throws CoreException {
		List<String> toskip = Arrays.asList((membersToSkip + ",createdDate,changedDate").split(","));
		List<String> tokeep = Arrays.asList((membersToKeep + ",System.owner,System.changedBy").split(","));
		List<String> revAssoc = Arrays.asList(reverseAssociations.split(","));
		List<String> skipEntities = Arrays.asList(excludeEntities.split(","));
		List<String> skipModules = Arrays.asList(excludeModules.split(","));
		Map<IMendixIdentifier, IMendixIdentifier> mappedIDs = new HashMap<IMendixIdentifier, IMendixIdentifier>();
		duplicate(c, source, target, toskip, tokeep, revAssoc, skipEntities, skipModules, mappedIDs);
	}

	private static void duplicate(IContext ctx, IMendixObject src, IMendixObject tar,
		List<String> toskip, List<String> tokeep, List<String> revAssoc,
		List<String> skipEntities, List<String> skipModules,
		Map<IMendixIdentifier, IMendixIdentifier> mappedObjects) throws CoreException {
		mappedObjects.put(src.getId(), tar.getId());

		Map<String, ? extends IMendixObjectMember<?>> members = src.getMembers(ctx);
		String type = src.getType() + "/";

		for (var entry : members.entrySet()) {
			String key = entry.getKey();
			if (!toskip.contains(key) && !toskip.contains(type + key)) {
				IMendixObjectMember<?> m = entry.getValue();
				if (m.isVirtual() || m instanceof MendixAutoNumber) {
					continue;
				}

				boolean keep = tokeep.contains(key) || tokeep.contains(type + key);

				if (m instanceof MendixObjectReference && !keep && m.getValue(ctx) != null) {
					IMendixObject o = Core.retrieveId(ctx, ((MendixObjectReference) m).getValue(ctx));
					IMendixIdentifier refObj = getCloneOfObject(ctx, o, toskip, tokeep, revAssoc, skipEntities, skipModules, mappedObjects);
					tar.setValue(ctx, key, refObj);
				} else if (m instanceof MendixObjectReferenceSet && !keep && m.getValue(ctx) != null) {
					MendixObjectReferenceSet rs = (MendixObjectReferenceSet) m;
					List<IMendixIdentifier> res = new ArrayList<IMendixIdentifier>();
					for (IMendixIdentifier item : rs.getValue(ctx)) {
						IMendixObject o = Core.retrieveId(ctx, item);
						IMendixIdentifier refObj = getCloneOfObject(ctx, o, toskip, tokeep, revAssoc, skipEntities, skipModules, mappedObjects);
						res.add(refObj);
					}
					tar.setValue(ctx, key, res);
				} else if (m instanceof MendixAutoNumber) //skip autonumbers! Ticket 14893
				{
					// do nothing
				} else if ("__UUID__".equals(key) && (isFileDocument(src) || isFileDocument(tar))) {
					// do nothing
				} else {
					tar.setValue(ctx, key, m.getValue(ctx));
				}
			}
		}
		Core.commitWithoutEvents(ctx, tar);
		duplicateReverseAssociations(ctx, src, tar, toskip, tokeep, revAssoc, skipEntities, skipModules, mappedObjects);
	}

	private static IMendixIdentifier getCloneOfObject(IContext ctx, IMendixObject src,
		List<String> toskip, List<String> tokeep, List<String> revAssoc,
		List<String> skipEntities, List<String> skipModules,
		Map<IMendixIdentifier, IMendixIdentifier> mappedObjects) throws CoreException {
		String objType = src.getMetaObject().getName();
		String modName = src.getMetaObject().getModuleName();

		// if object is already being cloned, return ref to clone
		if (mappedObjects.containsKey(src.getId())) {
			return mappedObjects.get(src.getId());
			// if object should be skipped based on module or entity, return source object
		} else if (skipEntities.contains(objType) || skipModules.contains(modName)) {
			return src.getId();
			// if not already being cloned, create clone
		} else {
			IMendixObject clone = Core.instantiate(ctx, src.getType());
			duplicate(ctx, src, clone, toskip, tokeep, revAssoc, skipEntities, skipModules, mappedObjects);
			return clone.getId();
		}
	}

	private static void duplicateReverseAssociations(IContext ctx, IMendixObject src, IMendixObject tar,
		List<String> toskip, List<String> tokeep, List<String> revAssocs,
		List<String> skipEntities, List<String> skipModules,
		Map<IMendixIdentifier, IMendixIdentifier> mappedObjects) throws CoreException {
		for (String fullAssocName : revAssocs) {
			String[] parts = fullAssocName.split("/");

			if (parts.length != 1 && parts.length != 3) //specifying entity has no meaning anymore, but remain backward compatible.
			{
				throw new IllegalArgumentException("Reverse association is not defined correctly, please mention the relation name only: '" + fullAssocName + "'");
			}

			String assocname = parts.length == 3 ? parts[1] : parts[0]; //support length 3 for backward compatibility

			IMetaAssociation massoc = src.getMetaObject().getDeclaredMetaAssociationChild(assocname);

			if (massoc != null) {
				IMetaObject relationParent = massoc.getParent();
				// if the parent is in the exclude list, we can't clone the parent, and setting the
				// references to the newly cloned target object will screw up the source data.
				if (skipEntities.contains(relationParent.getName()) || skipModules.contains(relationParent.getModuleName())) {
					throw new IllegalArgumentException("A reverse reference has been specified that starts at an entity in the exclude list, this is not possible to clone: '" + fullAssocName + "'");
				}

				//MWE: what to do with reverse reference sets? -> to avoid spam creating objects on
				//reverse references, do not support referenceset (todo: we could keep a map of converted guids and reuse that!)
				if (massoc.getType() == AssociationType.REFERENCESET) {
					throw new IllegalArgumentException("It is not possible to clone reverse referencesets: '" + fullAssocName + "'");
				}

				List<IMendixObject> objs = Core.createXPathQuery(String.format("//%s[%s=$value]", relationParent.getName(), assocname))
					.setVariable("value", src)
					.execute(ctx);

				for (IMendixObject obj : objs) {
					@SuppressWarnings("unused") // object is unused on purpose
					IMendixIdentifier refObj = getCloneOfObject(ctx, obj, toskip, tokeep, revAssocs, skipEntities, skipModules, mappedObjects);
					// setting reference explicitly is not necessary, this has been done in the
					// duplicate() call.
				}
			}
		}
	}

	public static Boolean commitWithoutEvents(IContext context, IMendixObject subject) throws CoreException {
		Core.commitWithoutEvents(context, subject);
		return true;
	}

	public static String getValueOfPath(IContext context, IMendixObject substitute, String fullpath, String datetimeformat) throws Exception {
		String[] path = fullpath.split("/");
		if (path.length == 1) {
			IMendixObjectMember<?> member = substitute.getMember(context, path[0]);

			//special case, see ticket 9135, format datetime.
			if (member instanceof MendixDateTime) {
				Date time = ((MendixDateTime) member).getValue(context);
				if (time == null) {
					return "";
				}
				String f = datetimeformat != null && !datetimeformat.isEmpty() ? datetimeformat : "EEE dd MMM yyyy, HH:mm";
				return new SimpleDateFormat(f).format(time);
			}

			if (member instanceof MendixEnum) {
				String value = member.parseValueToString(context);
				if (value == null || value.isEmpty()) {
					return "";
				}

				IMetaEnumeration enumeration = ((MendixEnum) member).getEnumeration();
				IMetaEnumValue evalue = enumeration.getEnumValues().get(value);
				return Core.getInternationalizedString(context, evalue.getI18NCaptionKey());
			}
			//default
			return member.parseValueToString(context);
		} else if (path.length == 0) {
			throw new Exception("communitycommons.ORM.getValueOfPath: Unexpected end of path.");
		} else {
			IMendixObjectMember<?> member = substitute.getMember(context, path[0]);
			if (member instanceof MendixObjectReference) {
				MendixObjectReference ref = (MendixObjectReference) member;
				IMendixIdentifier id = ref.getValue(context);
				if (id == null) {
					return "";
				}
				IMendixObject obj = Core.retrieveId(context, id);
				if (obj == null) {
					return "";
				}
				return getValueOfPath(context, obj, fullpath.substring(fullpath.indexOf("/") + 1), datetimeformat);
			} else if (member instanceof MendixObjectReferenceSet) {
				MendixObjectReferenceSet ref = (MendixObjectReferenceSet) member;
				List<IMendixIdentifier> ids = ref.getValue(context);
				if (ids == null) {
					return "";
				}
				StringBuilder res = new StringBuilder();
				for (IMendixIdentifier id : ids) {
					if (id == null) {
						continue;
					}
					IMendixObject obj = Core.retrieveId(context, id);
					if (obj == null) {
						continue;
					}
					res.append(", ");
					res.append(getValueOfPath(context, obj, fullpath.substring(fullpath.indexOf("/") + 1), datetimeformat));
				}
				return res.length() > 1 ? res.toString().substring(2) : "";
			} else {
				throw new Exception("communitycommons.ORM.getValueOfPath: Not a valid reference: '" + path[0] + "' in '" + fullpath + "'");
			}
		}
	}

	private static boolean isFileDocument(IMendixObject object) {
		return Core.isSubClassOf(Core.getMetaObject(FileDocument.entityName), object.getMetaObject());
	}

	public static Boolean cloneObject(IContext c, IMendixObject source,
		IMendixObject target, Boolean withAssociations) {
		Map<String, ? extends IMendixObjectMember<?>> members = source.getMembers(c);

		for (var entry : members.entrySet()) {
			IMendixObjectMember<?> m = entry.getValue();
			if (m.isVirtual()) {
				continue;
			}
			if (m instanceof MendixAutoNumber) {
				continue;
			}
			if ("__UUID__".equals(m.getName()) && (isFileDocument(source) || isFileDocument(target))) {
				continue;
            }
			if (withAssociations || ((!(m instanceof MendixObjectReference) && !(m instanceof MendixObjectReferenceSet) && !(m instanceof MendixAutoNumber)))) {
				target.setValue(c, entry.getKey(), m.getValue(c));
			}
		}
		return true;
	}

	public static IMendixObject firstWhere(IContext c, String entityName,
		Object member, String value) throws CoreException {
		List<IMendixObject> items = Core.retrieveXPathQuery(c, String.format("//%s[%s =  '%s']", entityName, member, value), 1, 0, new HashMap<String, String>());
		if (items == null || items.size() == 0) {
			return null;
		}
		return items.get(0);
	}

	public static IMendixObject getLastChangedByUser(IContext context,
		IMendixObject thing) throws CoreException {
		if (thing == null || !thing.hasChangedByAttribute()) {
			return null;
		}

		IMendixIdentifier itemId = thing.getChangedBy(context);
		if (itemId == null) {
			return null;
		}

		return Core.retrieveId(context, itemId);
	}

	public static IMendixObject getCreatedByUser(IContext context,
		IMendixObject thing) throws CoreException {
		if (thing == null || !thing.hasOwnerAttribute()) {
			return null;
		}

		IMendixIdentifier itemId = thing.getOwner(context);
		if (itemId == null) {
			return null;
		}

		return Core.retrieveId(context, itemId);
	}

	public static void commitSilent(IContext c, IMendixObject mendixObject) {
		try {
			Core.commit(c, mendixObject);
		} catch (CoreException e) {
			throw new RuntimeException(e);
		}
	}

	public static void copyAttributes(IContext context, IMendixObject source, IMendixObject target) {
		if (source == null) {
			throw new IllegalStateException("source is null");
		}
		if (target == null) {
			throw new IllegalStateException("target is null");
		}

		for (IMetaPrimitive e : target.getMetaObject().getMetaPrimitives()) {
			if (!source.hasMember(e.getName())) {
				continue;
			}
			if (e.isVirtual() || e.getType() == PrimitiveType.AutoNumber) {
				continue;
			}
			if ("__UUID__".equals(e.getName()) && (isFileDocument(source) || isFileDocument(target))) {
				continue;
			}

			target.setValue(context, e.getName(), source.getValue(context, e.getName()));
		}
	}
}
