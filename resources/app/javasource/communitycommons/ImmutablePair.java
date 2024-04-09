package communitycommons;

import org.apache.commons.lang3.builder.HashCodeBuilder;


public class ImmutablePair<T, U>
{
	public static <T, U> ImmutablePair<T, U> of(T left, U right) {
		return new ImmutablePair<T, U>(left, right);
	}

	private final T	left;
	private final U	right;

	private ImmutablePair(T left, U right) {
		if (left == null)
			throw new IllegalArgumentException("Left is NULL");
		if (right == null)
			throw new IllegalArgumentException("Right is NULL");
		
		this.left = left;
		this.right = right;
	}
	
	public T getLeft() {
		return left;
	}
	
	public U getRight() {
		return right;
	}
	
	@Override
	public String toString() {
		return "<" + left.toString()+ "," + right.toString() + ">";
	}
	
	@Override
	public boolean equals(Object other) {
		if (!(other instanceof ImmutablePair<?,?>))
			return false;
		
		if (this == other)
			return true;
		
		ImmutablePair<?,?> o = (ImmutablePair<?, ?>) other;
		return left.equals(o.getLeft()) && right.equals(o.getRight());
	}
	
	@Override
	public int hashCode() {
		return new HashCodeBuilder(19, 85)
			.append(left)
			.append(right)
			.toHashCode();
	}
	
}
