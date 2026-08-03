/** Преобразование в систему координат экрана */
export function translate(cs, pr)
{
  for (let i = 0; i < cs.length; i += 2)
  {
    cs[i] = (cs[i] - pr.left) * pr.scale;
    cs[i + 1] = (pr.top - cs[i + 1]) * pr.scale;
  }
}

/** Преобразование в систему координат экрана для описывающего прямоугольника */
export function translateRect(r, pr)
{
  r.left = (r.left - pr.left) * pr.scale;
  r.right = (r.right - pr.left) * pr.scale;

  r.bottom = (pr.top - r.bottom) * pr.scale;
  r.top = (pr.top - r.top) * pr.scale;
}

/** Преобразование в систему координат экрана для текста */
export function translateText(txt, pr)
{
  translate(txt.coords, pr);
  translateRect(txt.rect, pr);
  txt.x = (txt.x - pr.left) * pr.scale;
  txt.y = (pr.top - txt.y) * pr.scale;
}

/** Вычислить описывающий прямоугольник для координат */
export function calcRect(coords)
{

  let left = coords[0];
  let bottom = coords[1];

  let right = coords[0];
  let top = coords[1];

  for (let i = 2; i < coords.length; i += 2)
  {
    const x = coords[i];
    const y = coords[i + 1];

    if (left > x) left = x;
    if (bottom > y) bottom = y;

    if (right < x) right = x;
    if (top < y) top = y;

  }

  return { left, bottom, right, top };
}

/** Вычислить максимальный отрезок */
export function calcMaxLen(coords)
{
  let result = 0;
  let x1 = coords[0];
  let y1 = coords[1];
  let max = 0;


  for (let i = 2; i < coords.length; i += 2)
  {
    let x2 = coords[i];
    let y2 = coords[i + 1];

    const dx = x2 - x1;
    const dy = y2 - y1;

    const l = Math.hypot(dx, dy);
    if (max < l)
    {
      max = l;
      result = i - 2;
    }

  }

  return result;
}

/** Удаление точек которые не будут отображаться */
export function optimize(mas, l = 1)
{
  const count = mas.length;
  if (count < 5) return mas;

  const coords = [];
  let lastCoordX1 = mas[0];
  let lastCoordY1 = mas[1];
  let lastCoordX2 = mas[2];
  let lastCoordY2 = mas[3];
  coords.push(lastCoordX1, lastCoordY1);


  const lSq = l * l;
  for (let i = 4; i < count; i += 2)
    if (!isPointOnLine(lastCoordX1, lastCoordY1, lastCoordX2, lastCoordY2, mas[i], mas[i + 1], lSq))
    {
      lastCoordX1 = mas[i - 2];
      lastCoordY1 = mas[i - 1];

      lastCoordX2 = mas[i];
      lastCoordY2 = mas[i + 1];
      coords.push(lastCoordX1, lastCoordY1);
    }

  coords.push(mas[count - 2], mas[count - 1]);

  return coords;
}

/** Находится ли следующая точка на линии с определённым допуском */
export function isPointOnLine(pX1, pY1, pX2, pY2, pX, pY, distance)
{
  const a = pX - pX1;
  const b = pY - pY1;
  const c = pX2 - pX1;
  const d = pY2 - pY1;

  const lenSq = c * c + d * d;

  let dx, dy
  if (lenSq === 0)
  {
    // Точки совпадают — расстояние до точки
    dx = pX - pX1;
    dy = pY - pX1;
    return dx * dx + dy * dy < distance;
  }

  const param = (a * c + b * d) / lenSq;

  let xx, yy
  if (param < 0)
  {
    xx = pX1
    yy = pY1
  }
  else if (param > 1)
  {
    xx = pX2
    yy = pY2
  }
  else
  {
    xx = pX1 + param * c
    yy = pY1 + param * d
  }

  dx = pX - xx
  dy = pY - yy
  return dx * dx + dy * dy < distance
}

/** Преобразование из коэффициента в масштаб */
export function scale2Mashtab(scale)
{
  return (144 * 1000 / 25.4) / scale;
}

/** Преобразование из масштаба в коэффициент */
export function mashtab2Scale(scale)
{
  return scale * (144 * 1000 / 25.4);
}
