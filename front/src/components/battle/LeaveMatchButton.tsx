type Props = {
  disabled?: boolean;
  onLeave: () => void;
};

export function LeaveMatchButton({ disabled = false, onLeave }: Props) {
  return (
    <button type="button" className="battle-leave-button" onClick={onLeave} disabled={disabled}>
      ПОКИНУТЬ МАТЧ
    </button>
  );
}
