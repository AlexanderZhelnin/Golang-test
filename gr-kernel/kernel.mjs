import { optimize, translate } from "./calc.mjs";
import { clipPolygon } from "./polygon.mjs";
import { clipPolyline } from "./polyline.mjs";

/** Преобразование примитивов с учётом отсечения */
export function* clipPrimitives(l, pr, rect)
{
  for (const g of l.primitives)
  {
    if (
      g.rect.left >= rect.left &&
      g.rect.bottom >= rect.bottom &&
      g.rect.right <= rect.right &&
      g.rect.top <= rect.top
    )
      // Целиком лежит внутри прямоугольника
      yield { name: g.name, coords: [...g.coords] };
    else if (
      g.rect.left < rect.right &&
      g.rect.bottom < rect.top &&
      g.rect.right > rect.left &&
      g.rect.top > rect.bottom
    )
      // Необходимо отсекать
      switch (l.type)
      {
        case 1:
          for (const cs of clipPolyline(g, rect)) yield { name: g.name, coords: cs };
          break;
        case 2:
          {
            const cs = clipPolygon(g, rect);
            if (cs.length > 0) yield { name: g.name, coords: cs };
            break;
          }
      }
  }
}

/** Подготовка данных для отрисовки */
export function build(ls, pr, rect)
{
  const result = [];
  const mashtab = 1 / pr.scale;

  for (const l of ls)
  {
    const mas = [];

    result.push({ legendId: l.id, obrazes: mas });

    for (const obraz of clipPrimitives(l, pr, rect))
    {
      const csOpt = optimize(obraz.coords, mashtab);
      translate(csOpt, pr);
      mas.push({ name: obraz.name, coords: csOpt });
    }
  }

  return result;
}
