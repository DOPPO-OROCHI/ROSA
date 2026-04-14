import { resolveCardAssetVariantSrc } from "../../lib/api";
import type { BattleCardInMatch } from "./types";

type Props = {
  card: BattleCardInMatch;
  selected?: boolean;
};

export function HandCard({ card, selected = false }: Props) {
  return (
    <button type="button" className={`battle-hand-card ${selected ? "battle-hand-card--selected" : ""}`}>
      <img src={resolveCardAssetVariantSrc("battle", card.template_id, "view")} alt={card.template_id} />
    </button>
  );
}
