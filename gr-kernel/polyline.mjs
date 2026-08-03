/** Отсечение полилинии по прямоугольнику */
export function clipPolyline(g, rect)
{
  function clipLeft(coords, left)
  {
    const res = [];
    if (coords.length == 0) return res;

    let pl = [];

    let px1 = coords[0];
    let py1 = coords[1];

    if (px1 >= left)
    {
      pl.push(px1);
      pl.push(py1);
    }

    for (let i = 2; i < coords.length; i += 2)
    {
      let px2 = coords[i];
      let py2 = coords[i + 1];

      if (px1 >= left && px2 >= left)
      {
        pl.push(px2);
        pl.push(py2);
      }
      else if (px1 < left && px2 > left)
      {
        pl.push(left);
        pl.push((left - px1) * (py2 - py1) / (px2 - px1) + py1);

        pl.push(px2);
        pl.push(py2);
      }
      else if (px1 > left && px2 < left)
      {
        pl.push(left);
        pl.push((left - px1) * (py2 - py1) / (px2 - px1) + py1);

        res.push(pl);
        pl = [];

      }
      px1 = px2;
      py1 = py2;
    }
    if (pl.length > 0) res.push(pl);
    return res;
  }
  function clipRight(coords, right)
  {
    const res = [];
    if (coords.length == 0) return res;

    let pl = [];

    let px1 = coords[0];
    let py1 = coords[1];

    if (px1 <= right)
    {
      pl.push(px1);
      pl.push(py1);
    }

    for (let i = 2; i < coords.length; i += 2)
    {
      let px2 = coords[i];
      let py2 = coords[i + 1];

      if (px1 <= right && px2 <= right)
      {
        pl.push(px2);
        pl.push(py2);
      }
      else if (px1 > right && px2 < right)
      {
        pl.push(right);
        pl.push((right - px1) * (py2 - py1) / (px2 - px1) + py1);

        pl.push(px2);
        pl.push(py2);
      }
      else if (px1 < right && px2 > right)
      {
        pl.push(right);
        pl.push((right - px1) * (py2 - py1) / (px2 - px1) + py1);

        res.push(pl);

        pl = [];
      }
      px1 = px2;
      py1 = py2;
    }
    if (pl.length > 0) res.push(pl);
    return res;
  }
  function clipBottom(coords, bottom)
  {
    const res = [];
    if (coords.length == 0) return res;

    let pl = [];

    let px1 = coords[0];
    let py1 = coords[1];

    if (py1 >= bottom)
    {
      pl.push(px1);
      pl.push(py1);
    }

    for (let i = 2; i < coords.length; i += 2)
    {
      let px2 = coords[i];
      let py2 = coords[i + 1];

      if (py1 >= bottom && py2 >= bottom)
      {
        pl.push(px2);
        pl.push(py2);
      }
      else if (py1 < bottom && py2 > bottom)
      {
        pl.push((bottom - py1) * (px2 - px1) / (py2 - py1) + px1);
        pl.push(bottom);

        pl.push(px2);
        pl.push(py2);
      }
      else if (py1 > bottom && py2 < bottom)
      {
        pl.push(px1);
        pl.push(py1);

        pl.push((bottom - py1) * (px2 - px1) / (py2 - py1) + px1);
        pl.push(bottom);

        res.push(pl);

        pl = [];
      }
      px1 = px2;
      py1 = py2;
    }
    if (pl.length > 0) res.push(pl);
    return res;
  }
  function clipTop(coords, top)
  {
    const res = [];
    if (coords.length == 0) return res;

    let pl = [];

    let px1 = coords[0];
    let py1 = coords[1];

    if (py1 <= top)
    {
      pl.push(px1);
      pl.push(py1);
    }


    for (let i = 2; i < coords.length; i += 2)
    {
      let px2 = coords[i];
      let py2 = coords[i + 1];

      if (py1 <= top && py2 <= top)
      {
        pl.push(px2);
        pl.push(py2);
      }
      else if (py1 < top && py2 > top)
      {
        pl.push(px1);
        pl.push(py1);

        pl.push((top - py1) * (px2 - px1) / (py2 - py1) + px1);
        pl.push(top);
        res.push(pl);
        pl = [];
      }
      else if (py1 > top && py2 < top)
      {
        pl.push((top - py1) * (px2 - px1) / (py2 - py1) + px1);
        pl.push(top);

        pl.push(px2);
        pl.push(py2);
      }

      px1 = px2;
      py1 = py2;
    }
    if (pl.length > 0) res.push(pl);
    return res;
  }

  let res = [];

  if (g.rect.left < rect.left)
    res = clipLeft(g.coords, rect.left);
  else
    res.push([...g.coords]);

  if (g.rect.bottom < rect.bottom)
  {
    const tmp = [];
    for (const cs of res)
      tmp.push(...clipBottom(cs, rect.bottom));
    res = tmp;
  }

  if (g.rect.right > rect.right)
  {
    const tmp = [];
    for (const cs of res)
      tmp.push(...clipRight(cs, rect.right));
    res = tmp;
  }

  if (g.rect.top > rect.top)
  {
    const tmp = [];
    for (const cs of res)
      tmp.push(...clipTop(cs, rect.top));
    res = tmp;
  }

  return res;
}
