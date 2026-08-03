/** Отсечение полигона по прямоугольнику */
export function clipPolygon(g, rect)
{
  let curIndex = 0;

  function getNextIndex(coords)
  {
    curIndex += 2;
    if (curIndex >= coords.length) curIndex = 0;
    return curIndex;

  }
  function clipLeft(coords, left)
  {
    if (coords.length == 0) return coords;

    curIndex = 0;
    let pl = [];

    let index = getNextIndex(coords);

    let px1 = coords[index];
    let py1 = coords[index + 1];

    if (px1 >= left)
    {
      pl.push(px1);
      pl.push(py1);
    }

    const len = coords.length / 2;
    for (let i = 1; i <= len; i++)
    {
      index = getNextIndex(coords);
      let px2 = coords[index];
      let py2 = coords[index + 1];

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
      }
      px1 = px2;
      py1 = py2;
    }

    return pl;
  }
  function clipRight(coords, right)
  {
    if (coords.length == 0) return coords;

    curIndex = 0;

    let pl = [];

    let index = getNextIndex(coords);

    let px1 = coords[index];
    let py1 = coords[index + 1];

    if (px1 <= right)
    {
      pl.push(px1);
      pl.push(py1);
    }
    const len = coords.length / 2;

    for (let i = 0; i < len; i++)
    {
      index = getNextIndex(coords);

      let px2 = coords[index];
      let py2 = coords[index + 1];

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
      }
      px1 = px2;
      py1 = py2;
    }

    return pl;
  }
  function clipBottom(coords, bottom)
  {
    if (coords.length == 0) return coords;

    curIndex = 0;

    let pl = [];

    let index = getNextIndex(coords);

    let px1 = coords[index];
    let py1 = coords[index + 1];

    if (py1 >= bottom)
    {
      pl.push(px1);
      pl.push(py1);
    }

    const len = coords.length / 2;
    for (let i = 0; i < len; i++)
    {
      index = getNextIndex(coords);
      let px2 = coords[index];
      let py2 = coords[index + 1];

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
        pl.push((bottom - py1) * (px2 - px1) / (py2 - py1) + px1);
        pl.push(bottom);
      }
      px1 = px2;
      py1 = py2;
    }

    return pl;
  }
  function clipTop(coords, top)
  {
    if (coords.length == 0) return coords;
    curIndex = 0;


    let pl = [];

    let index = getNextIndex(coords);

    let px1 = coords[index];
    let py1 = coords[index + 1];

    if (py1 <= top)
    {
      pl.push(px1);
      pl.push(py1);
    }

    const len = coords.length / 2;
    for (let i = 0; i < len; i++)
    {
      index = getNextIndex(coords);
      let px2 = coords[index];
      let py2 = coords[index + 1];

      if (py1 <= top && py2 <= top)
      {
        pl.push(px2);
        pl.push(py2);
      }
      else if (py1 < top && py2 > top)
      {
        pl.push((top - py1) * (px2 - px1) / (py2 - py1) + px1);
        pl.push(top);
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

    return pl;
  }

  let res = [];

  if (g.rect.left < rect.left)
    res = clipLeft(g.coords, rect.left);
  else
    res.push(...g.coords);

  if (g.rect.bottom < rect.bottom)
    res = clipBottom(res, rect.bottom);


  if (g.rect.right > rect.right)
    res = clipRight(res, rect.right);

  if (g.rect.top > rect.top)
    res = clipTop(res, rect.top);

  return res;
}
